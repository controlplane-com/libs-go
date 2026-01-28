package replication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"go.uber.org/zap"
)

func StartReplication(c *Config) error {
	if c == nil {
		return fmt.Errorf("replication config cannot be nil")
	}
	conn, err := replicationConnection(c)
	if err != nil {
		return err
	}

	lsn, err := getRestartLsn(c)
	if err != nil {
		return err
	}
	addTablesArg := fmt.Sprintf("\"add-tables\" '%s'", strings.Join(c.Tables(), ","))
	var args []string
	args = []string{"\"pretty-print\" 'true'", addTablesArg}
	if err := pglogrepl.StartReplication(context.Background(), conn.PgConn(), c.Slot, lsn, pglogrepl.StartReplicationOptions{PluginArgs: args}); err != nil {
		return fmt.Errorf("StartReplication failed: %v", err)
	}
	logging.Logger().Info("Logical replication started", zap.String("slot", c.Slot))
	return streamWal(c, conn.PgConn(), lsn)
}

func getRestartLsn(c *Config) (pglogrepl.LSN, error) {
	var restartLSN string
	conn, err := connection(c)
	if err != nil {
		return 0, err
	}
	err = conn.QueryRow(context.Background(), "SELECT restart_lsn FROM pg_replication_slots WHERE slot_name = $1", c.Slot).Scan(&restartLSN)
	if err != nil {
		return 0, fmt.Errorf("Failed to retrieve restart_lsn: %v\n", err)
	}
	lsn, err := pglogrepl.ParseLSN(restartLSN)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse restart_lsn: %v\n", err)
	}
	return lsn, nil
}

func streamWal(c *Config, conn *pgconn.PgConn, walPosition pglogrepl.LSN) error {
	nextStandbyMessageDeadline := time.Now().Add(c.WalAcknowledgementFrequency)
	for {
		if time.Now().After(nextStandbyMessageDeadline) {
			err := pglogrepl.SendStandbyStatusUpdate(context.Background(), conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: walPosition})
			if err != nil {
				return fmt.Errorf("SendStandbyStatusUpdate failed: %v", err)
			}
			nextStandbyMessageDeadline = time.Now().Add(c.WalAcknowledgementFrequency)
		}

		ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		rawMsg, err := conn.ReceiveMessage(ctx)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) {
				continue
			}
			return fmt.Errorf("ReceiveMessage failed: %v", err)
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			return fmt.Errorf("received Postgres WAL error: %+v", errMsg)
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			logging.Logger().Sugar().Warnf("Received unexpected message: %T\n", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				return fmt.Errorf("ParsePrimaryKeepaliveMessage failed: %v", err)
			}
			if pkm.ServerWALEnd > walPosition {
				walPosition = pkm.ServerWALEnd
			}
			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				return fmt.Errorf("ParseXLogData failed: %v", err)
			}

			var walData WalData
			if err := json.Unmarshal(xld.WALData, &walData); err != nil {
				logging.Logger().Sugar().Errorf("unable to parse wal2json change: %v", err)
				if xld.WALStart > walPosition {
					walPosition = xld.WALStart
				}
				continue
			}

			for _, change := range walData.Changes {
				if err := c.Destinations.HandleChange(change); err != nil {
					logger := logging.Logger().Sugar()
					logger.Errorf("error handling change for replication slot %s: %v", c.Slot, err)
					logger.Infof("wal2json data: \n%s\n", string(xld.WALData))
					if errors.Is(err, IgnoreMessageError) {
						continue
					}
					//Panic so we don't acknowledge WAL entries. We'll try again on the next reboot
					panic(err)
				}
			}
			if xld.WALStart > walPosition {
				walPosition = xld.WALStart
			}
		}
	}
}

func replicationConnection(config *Config) (*pgx.Conn, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable replication=database",
		config.Host, config.User, config.Password, config.Database, config.Port)
	return pgx.Connect(context.Background(), dsn)
}
func connection(config *Config) (*pgx.Conn, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Host, config.User, config.Password, config.Database, config.Port)
	return pgx.Connect(context.Background(), dsn)
}

package checkpoints

import (
	"github.com/controlplane-com/libs-go/pkg/database"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type PostgresqlCheckpointRepository struct {
	database.Repository
	connection database.Connection
}

// EnforceMinimumCheckpointTime sets all non-default checkpoints to the given minimum time, except
// the rows named in excludeNames. Excluded rows are the ones the metering engine advances itself:
// force-advancing one of those would silently cancel a checkpoint reset — the hours between the
// reset and the minimum would never be metered, with nothing failing.
func (r *PostgresqlCheckpointRepository) EnforceMinimumCheckpointTime(minimum time.Time, excludeNames []string) error {
	query := r.connection.Db().Table("checkpoints").
		Where("last_completed_time < ? AND name != ? AND job != ?", minimum, DefaultCheckpointName, DefaultCheckpointName)
	if len(excludeNames) > 0 {
		query = query.Where("name NOT IN ?", excludeNames)
	}
	return query.Update("last_completed_time", minimum).Error
}

func NewPostgresqlCheckpointRepository() CheckpointRepository {
	return &PostgresqlCheckpointRepository{}
}

func (r *PostgresqlCheckpointRepository) Initialize(connection database.Connection) error {
	if err := connection.Initialize(); err != nil {
		return err
	}
	r.connection = connection
	return nil
}

func (r *PostgresqlCheckpointRepository) QueryCheckpoints(query *CheckpointQuery) ([]*Checkpoint, error) {
	return r.queryCheckpoints(r.connection.Db(), query)
}

func (r *PostgresqlCheckpointRepository) GetCheckpoint(query *CheckpointQuery) (*Checkpoint, error) {
	return r.getCheckpoint(r.connection.Db(), query)
}

func (r *PostgresqlCheckpointRepository) ListCheckpoints() ([]*Checkpoint, error) {
	return r.queryCheckpoints(r.connection.DbRo(), nil)
}

func (r *PostgresqlCheckpointRepository) SetCheckpoint(checkpoint *Checkpoint) error {
	return r.setCheckpoint(r.connection.Db(), checkpoint)
}

func (r *PostgresqlCheckpointRepository) queryCheckpoints(db *gorm.DB, query *CheckpointQuery) ([]*Checkpoint, error) {
	var checkpoints []*Checkpoint
	if query != nil {
		db = r.addCheckpointWhereClause(db, query)
	}
	result := db.Find(&checkpoints)
	if result.Error != nil {
		return nil, result.Error
	}
	pipeline.MustMap(checkpoints, func(c *Checkpoint) *Checkpoint {
		c.LastCompletedTime = c.LastCompletedTime.UTC()
		return c
	})
	return checkpoints, nil
}

func (r *PostgresqlCheckpointRepository) getCheckpoint(db *gorm.DB, query *CheckpointQuery) (*Checkpoint, error) {
	checkpoints, err := r.queryCheckpoints(db, query)
	if err != nil {
		return nil, err
	}
	if len(checkpoints) == 0 {
		name := query.Name
		job := query.Job
		if job != DefaultCheckpointName && name != DefaultCheckpointName {
			jobCheckpoint, err := r.getJobDefaultCheckpoint(db, job)
			if err != nil {
				return nil, err
			}
			if jobCheckpoint != nil {
				return jobCheckpoint, nil
			}
		}
		return r.getGlobalDefaultCheckpoint(db)
	}
	return checkpoints[0], nil
}

func (r *PostgresqlCheckpointRepository) getGlobalDefaultCheckpoint(db *gorm.DB) (*Checkpoint, error) {
	checkpoints, err := r.queryCheckpoints(db, &CheckpointQuery{Name: DefaultCheckpointName, Job: DefaultCheckpointName})
	if err != nil {
		return nil, err
	}
	if len(checkpoints) > 0 {
		return checkpoints[0], nil
	}
	return nil, nil
}

func (r *PostgresqlCheckpointRepository) getJobDefaultCheckpoint(db *gorm.DB, job string) (*Checkpoint, error) {
	var checkpoint = Checkpoint{
		Name: DefaultCheckpointName,
		Job:  job,
	}
	result := db.First(&checkpoint)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	checkpoint.LastCompletedTime = checkpoint.LastCompletedTime.UTC()
	return &checkpoint, nil
}

func (r *PostgresqlCheckpointRepository) setCheckpoint(db *gorm.DB, checkpoint *Checkpoint) error {
	result := db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "name"},
				{Name: "job"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"last_completed_time"}),
		}).
		Save(checkpoint)
	return result.Error
}

func (r *PostgresqlCheckpointRepository) setAllCheckpoints(db *gorm.DB, newCheckpointTime time.Time) error {
	result := db.
		Model(Checkpoint{}).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Update("last_completed_time", newCheckpointTime)
	return result.Error
}

func (r *PostgresqlCheckpointRepository) resetCheckpoints(db *gorm.DB, request *ResetCheckpointsRequest) error {
	result := r.addCheckpointWhereClause(
		db.Model(Checkpoint{}),
		&request.CheckpointQuery,
	).Update("last_completed_time", request.NewCheckpointTime)
	return result.Error
}

func (r *PostgresqlCheckpointRepository) addCheckpointWhereClause(db *gorm.DB, query *CheckpointQuery) *gorm.DB {
	var where []string
	var args []any
	if query.Name != "" {
		where = append(where, "name = ?")
		args = append(args, query.Name)
	}
	if query.Job != "" {
		where = append(where, "job = ?")
		args = append(args, query.Job)
	}
	return db.Where(strings.Join(where, " AND "), args...)
}

func (r *PostgresqlCheckpointRepository) ResetAllCheckpoints(newCheckpointTime time.Time) error {
	return r.setAllCheckpoints(r.connection.Db(), newCheckpointTime)
}

func (r *PostgresqlCheckpointRepository) ResetCheckpoints(request *ResetCheckpointsRequest) error {
	if request == nil {
		request = &ResetCheckpointsRequest{}
	}
	return r.connection.Db().Transaction(func(tx *gorm.DB) error {
		if request.Job != DefaultCheckpointName && request.Name != DefaultCheckpointName {
			jobDefault := Checkpoint{}
			result := tx.Where("job = ? AND name = ? AND last_completed_time > ?", request.Job, DefaultCheckpointName, request.NewCheckpointTime).First(&jobDefault)
			if result.RowsAffected > 0 {
				tx.Model(jobDefault).Update("last_completed_time", request.NewCheckpointTime)
			}
		}
		if request.Job != DefaultCheckpointName || request.Name != DefaultCheckpointName {
			globalDefault := Checkpoint{}
			result := tx.Where("job = ? AND name = ? AND last_completed_time > ?", DefaultCheckpointName, DefaultCheckpointName, request.NewCheckpointTime).First(&globalDefault)
			if result.RowsAffected > 0 {
				tx.Model(globalDefault).Update("last_completed_time", request.NewCheckpointTime)
			}
		}
		return r.resetCheckpoints(tx, request)
	})
}

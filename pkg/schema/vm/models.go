/* hand-maintained — canonical source of truth (nodelibs/schema/src/vm.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go.
 *
 * Object-level relations/customs and the SecretLink / VmUser leaf rules all live in
 * vm/validate.go, via the Validator interface. */

package vm

// SecretLink is LocalLink+ToKind("secret") (its rule is in validate.go).
type SecretLink string

// VmUser is String().Max(32).Regex(vmUserRe) (its rule is in validate.go).
type VmUser string

// The following nominal leaf types are self-validating via the Validator interface;
// their ValidationRule methods (and the named regexes/customs) live in validate.go.

// VmImageLink is the strongly-typed replacement for the `imageLink` tag on
// VmSpecBootDiskSourceOci.Image (base ImageLink; the image kind check is embedded).
type VmImageLink string

// VmHttpUrl replaces `custom=vmHttpUrl` on VmSpecBootDiskSourceHttp.Url.
type VmHttpUrl string

// VmChecksum replaces `pattern=vmChecksum` on VmSpecBootDiskSourceHttp.Checksum.
type VmChecksum string

// VmVolumeSetUri replaces `pattern=vmVolumeSetUri` on VmSpecBootDiskPersist.VolumeSet.
type VmVolumeSetUri string

// VmNetName replaces `pattern=vmNetName` on VmSpecNetworks.Name.
type VmNetName string

// VmHost replaces `pattern=vmHost` on VmSpec.Hostname / VmSpec.Subdomain.
type VmHost string

// VmBase64 replaces `custom=vmBase64` on VmSpecCloudInit.UserDataBase64.
type VmBase64 string

type VmSpecBootDiskSourceOci struct {
	Image VmImageLink `json:"image" validate:"required"`
}

type VmSpecBootDiskSourceHttp struct {
	Url      VmHttpUrl  `json:"url" validate:"required"`
	Checksum VmChecksum `json:"checksum,omitzero" validate:"max=160"`
}

type VmSpecBootDiskSource struct {
	Oci  *VmSpecBootDiskSourceOci  `json:"oci,omitzero"`
	Http *VmSpecBootDiskSourceHttp `json:"http,omitzero"`
}

type VmSpecBootDiskPersist struct {
	VolumeSet VmVolumeSetUri `json:"volumeSet" validate:"required"`
}

type VmSpecBootDiskBus string

const (
	VmSpecBootDiskBusVirtio VmSpecBootDiskBus = "virtio"
	VmSpecBootDiskBusSata   VmSpecBootDiskBus = "sata"
	VmSpecBootDiskBusScsi   VmSpecBootDiskBus = "scsi"
)

func (VmSpecBootDiskBus) EnumValues() []string { return []string{"virtio", "sata", "scsi"} }

type VmSpecBootDisk struct {
	Source    *VmSpecBootDiskSource  `json:"source,omitzero"`
	Persist   *VmSpecBootDiskPersist `json:"persist,omitzero"`
	Bus       VmSpecBootDiskBus      `json:"bus,omitzero" validate:"default=virtio"`
	BootOrder *float64               `json:"bootOrder,omitzero" validate:"integer,min=1,max=16,default=1"`
}

type VmSpecCpu struct {
	Sockets *float64 `json:"sockets,omitzero" validate:"integer,min=1,max=32"`
	Threads *float64 `json:"threads,omitzero" validate:"integer,min=1,max=8"`
}

type VmSpecFirmwareBootloader string

const (
	VmSpecFirmwareBootloaderBios VmSpecFirmwareBootloader = "bios"
	VmSpecFirmwareBootloaderEfi  VmSpecFirmwareBootloader = "efi"
)

func (VmSpecFirmwareBootloader) EnumValues() []string { return []string{"bios", "efi"} }

type VmSpecFirmwareSmbios struct {
	Manufacturer string `json:"manufacturer,omitzero" validate:"max=64"`
	Product      string `json:"product,omitzero" validate:"max=64"`
	Version      string `json:"version,omitzero" validate:"max=64"`
	Sku          string `json:"sku,omitzero" validate:"max=64"`
	Family       string `json:"family,omitzero" validate:"max=64"`
}

type VmSpecFirmware struct {
	Bootloader VmSpecFirmwareBootloader `json:"bootloader,omitzero" validate:"default=efi"`
	SecureBoot *bool                    `json:"secureBoot,omitzero" validate:"default=false"`
	Uuid       string                   `json:"uuid,omitzero" validate:"uuid"`
	Serial     string                   `json:"serial,omitzero" validate:"max=64"`
	Smbios     *VmSpecFirmwareSmbios    `json:"smbios,omitzero"`
}

type VmSpecGuestOs string

const (
	VmSpecGuestOsLinux   VmSpecGuestOs = "linux"
	VmSpecGuestOsWindows VmSpecGuestOs = "windows"
)

func (VmSpecGuestOs) EnumValues() []string { return []string{"linux", "windows"} }

type VmSpecNetworks struct {
	Name VmNetName `json:"name,omitzero" validate:"max=15,default=default"`
}

type VmSpecCloudInit struct {
	UserData            string       `json:"userData,omitzero" validate:"max=16384"`
	UserDataBase64      VmBase64     `json:"userDataBase64,omitzero" validate:"max=22000"`
	UserDataSecret      SecretLink   `json:"userDataSecret,omitzero"`
	SshPublicKeySecrets []SecretLink `json:"sshPublicKeySecrets,omitzero" validate:"max=8"`
}

type VmSpecAccessCredentialsDeliveryMethod string

const (
	VmSpecAccessCredentialsDeliveryMethodQemuGuestAgent VmSpecAccessCredentialsDeliveryMethod = "qemuGuestAgent"
	VmSpecAccessCredentialsDeliveryMethodConfigDrive    VmSpecAccessCredentialsDeliveryMethod = "configDrive"
)

func (VmSpecAccessCredentialsDeliveryMethod) EnumValues() []string {
	return []string{"qemuGuestAgent", "configDrive"}
}

type VmSpecAccessCredentials struct {
	SshPublicKeySecret SecretLink                            `json:"sshPublicKeySecret" validate:"required"`
	Users              []VmUser                              `json:"users" validate:"min=1,max=16,required"`
	DeliveryMethod     VmSpecAccessCredentialsDeliveryMethod `json:"deliveryMethod,omitzero" validate:"default=qemuGuestAgent"`
}

type VmSpecRunStrategy string

const (
	VmSpecRunStrategyAlways         VmSpecRunStrategy = "Always"
	VmSpecRunStrategyRerunOnFailure VmSpecRunStrategy = "RerunOnFailure"
	VmSpecRunStrategyManual         VmSpecRunStrategy = "Manual"
	VmSpecRunStrategyHalted         VmSpecRunStrategy = "Halted"
)

func (VmSpecRunStrategy) EnumValues() []string {
	return []string{"Always", "RerunOnFailure", "Manual", "Halted"}
}

type VmSpecClock struct {
	Timezone string `json:"timezone,omitzero" validate:"max=64,default=UTC"`
}

type VmSpec struct {
	BootDisk          *VmSpecBootDisk           `json:"bootDisk,omitzero" validate:"default={}"`
	Cpu               *VmSpecCpu                `json:"cpu,omitzero"`
	Firmware          *VmSpecFirmware           `json:"firmware,omitzero" validate:"default={}"`
	GuestOS           VmSpecGuestOs             `json:"guestOS,omitzero" validate:"default=linux"`
	Networks          []VmSpecNetworks          `json:"networks,omitzero" validate:"max=1"`
	CloudInit         *VmSpecCloudInit          `json:"cloudInit,omitzero"`
	AccessCredentials []VmSpecAccessCredentials `json:"accessCredentials,omitzero" validate:"max=8"`
	RunStrategy       VmSpecRunStrategy         `json:"runStrategy,omitzero" validate:"default=Always"`
	Clock             *VmSpecClock              `json:"clock,omitzero" validate:"default={}"`
	Hostname          VmHost                    `json:"hostname,omitzero" validate:"max=63"`
	Subdomain         VmHost                    `json:"subdomain,omitzero" validate:"max=63"`
}

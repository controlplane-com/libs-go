/* auto-generated */

package vm

type VmSpecBootDiskSourceOci struct {
	Image string `json:"image"`
}

type VmSpecBootDiskSourceHttp struct {
	Url      string `json:"url"`
	Checksum string `json:"checksum,omitempty"`
}

type VmSpecBootDiskSource struct {
	Oci  VmSpecBootDiskSourceOci  `json:"oci,omitempty"`
	Http VmSpecBootDiskSourceHttp `json:"http,omitempty"`
}

type VmSpecBootDiskPersist struct {
	VolumeSet string `json:"volumeSet"`
}

type VmSpecBootDiskBus string

const (
	VmSpecBootDiskBusVirtio VmSpecBootDiskBus = "virtio"
	VmSpecBootDiskBusSata   VmSpecBootDiskBus = "sata"
	VmSpecBootDiskBusScsi   VmSpecBootDiskBus = "scsi"
)

type VmSpecBootDisk struct {
	Source    VmSpecBootDiskSource   `json:"source,omitempty"`
	Persist   *VmSpecBootDiskPersist `json:"persist,omitempty"`
	Bus       VmSpecBootDiskBus      `json:"bus,omitempty"`
	BootOrder *float32               `json:"bootOrder,omitempty"`
}

type VmSpecCpu struct {
	Sockets *float32 `json:"sockets,omitempty"`
	Threads *float32 `json:"threads,omitempty"`
}

type VmSpecFirmwareBootloader string

const (
	VmSpecFirmwareBootloaderBios VmSpecFirmwareBootloader = "bios"
	VmSpecFirmwareBootloaderEfi  VmSpecFirmwareBootloader = "efi"
)

type VmSpecFirmwareSmbios struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Product      string `json:"product,omitempty"`
	Version      string `json:"version,omitempty"`
	Sku          string `json:"sku,omitempty"`
	Family       string `json:"family,omitempty"`
}

type VmSpecFirmware struct {
	Bootloader VmSpecFirmwareBootloader `json:"bootloader,omitempty"`
	SecureBoot bool                     `json:"secureBoot,omitempty"`
	Uuid       string                   `json:"uuid,omitempty"`
	Serial     string                   `json:"serial,omitempty"`
	Smbios     VmSpecFirmwareSmbios     `json:"smbios,omitempty"`
}

type VmSpecGuestOs string

const (
	VmSpecGuestOsLinux   VmSpecGuestOs = "linux"
	VmSpecGuestOsWindows VmSpecGuestOs = "windows"
)

type VmSpecNetworks struct {
	Name string `json:"name,omitempty"`
}

type VmSpecCloudInit struct {
	UserData            string   `json:"userData,omitempty"`
	UserDataBase64      string   `json:"userDataBase64,omitempty"`
	UserDataSecret      string   `json:"userDataSecret,omitempty"`
	SshPublicKeySecrets []string `json:"sshPublicKeySecrets,omitempty"`
}

type VmSpecAccessCredentialsDeliveryMethod string

const (
	VmSpecAccessCredentialsDeliveryMethodQemuGuestAgent VmSpecAccessCredentialsDeliveryMethod = "qemuGuestAgent"
	VmSpecAccessCredentialsDeliveryMethodConfigDrive    VmSpecAccessCredentialsDeliveryMethod = "configDrive"
)

type VmSpecAccessCredentials struct {
	SshPublicKeySecret string                                `json:"sshPublicKeySecret"`
	Users              []string                              `json:"users"`
	DeliveryMethod     VmSpecAccessCredentialsDeliveryMethod `json:"deliveryMethod,omitempty"`
}

type VmSpecRunStrategy string

const (
	VmSpecRunStrategyAlways         VmSpecRunStrategy = "Always"
	VmSpecRunStrategyRerunOnFailure VmSpecRunStrategy = "RerunOnFailure"
	VmSpecRunStrategyManual         VmSpecRunStrategy = "Manual"
	VmSpecRunStrategyHalted         VmSpecRunStrategy = "Halted"
)

type VmSpecClock struct {
	Timezone string `json:"timezone,omitempty"`
}

type VmSpec struct {
	BootDisk          *VmSpecBootDisk           `json:"bootDisk,omitempty"`
	Cpu               VmSpecCpu                 `json:"cpu,omitempty"`
	Firmware          VmSpecFirmware            `json:"firmware,omitempty"`
	GuestOS           VmSpecGuestOs             `json:"guestOS,omitempty"`
	Networks          []VmSpecNetworks          `json:"networks,omitempty"`
	CloudInit         VmSpecCloudInit           `json:"cloudInit,omitempty"`
	AccessCredentials []VmSpecAccessCredentials `json:"accessCredentials,omitempty"`
	RunStrategy       VmSpecRunStrategy         `json:"runStrategy,omitempty"`
	Clock             VmSpecClock               `json:"clock,omitempty"`
	Hostname          string                    `json:"hostname,omitempty"`
	Subdomain         string                    `json:"subdomain,omitempty"`
}

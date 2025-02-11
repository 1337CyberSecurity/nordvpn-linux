package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	// ListenPID defines process id env key
	ListenPID = "LISTEN_PID"

	// ListenFDS defines systemDFile descriptors env key
	ListenFDS = "LISTEN_FDS"

	// ListenFDNames defines systemDFile descriptors names env key
	ListenFDNames = "LISTEN_FDNAMES"

	// Proto defines protocol to be used
	Proto = "unix"

	// TempDir defines temporary storage directory
	TempDir = "/tmp/"

	// RunDir defines default socket directory
	RunDir = "/run/nordvpn/"

	// LogPath defines where logs are located if systemd isn't used
	LogPath = "/var/log/nordvpn/"

	// NordvpnGroup that can access daemon socket
	NordvpnGroup = "nordvpn"

	// DaemonSocket defines system daemon socket file location
	DaemonSocket = RunDir + "nordvpnd.sock"

	// PermUserRWX user permission type to read write and execute
	PermUserRWX = 0700

	// PermUserRW user permission type to read and write
	PermUserRW = 0600

	// PermUserRWGroupRW permission type for user and group to read and write, everyone else - no access.
	PermUserRWGroupRW = 0660

	// PermUserRWGroupROthersR user permission type for user to read and write to it, everyone else can only read it.
	PermUserRWGroupROthersR = 0644

	// PermUserRWGroupROthersR allows user and group to read and write, other only read
	PermUserRWGroupRWOthersR = 0664

	// PermUserRWGroupROthersR user permission type for everyone to read and write to it.
	PermUserRWGroupRWOthersRW = 0666

	// PermUserRWXGroupRXOthersRX forbidding group and others to write to it
	PermUserRWXGroupRXOthersRX = 0755

	// ChattrExec is the chattr command executable name
	ChattrExec = "chattr"

	// Column is a tool to format data into columns for neater display in CLI
	ColumnExec = "column"

	// SttyExec is a tool to change or print CLI settings
	SttyExec = "stty"

	// SystemctlExec defines system controller executable
	SystemctlExec = "systemctl"

	// NetworkctlExec defines network controller executable
	NetworkctlExec = "networkctl"

	// ServerDateFormat defines api date format
	ServerDateFormat = "2006-01-02 15:04:05"

	// Fileshared defines filesharing daemon name
	Fileshared = "nordfileshared"

	// ConfigDirectory is used for configuration files storage. Hardcoded only for nordfileshared, in
	// other cases consider using os.UserConfigDir instead.
	ConfigDirectory = ".config"

	// FileshareHistoryFile is the storage file used by libdrop
	FileshareHistoryFile = "fileshare_history.db"
)

const (
	// UserDataPath defines path where user data is stored
	UserDataPath = "nordvpn/"

	// ResolvconfFilePath defines path to resolv.conf file for DNS
	ResolvconfFilePath = "/etc/resolv.conf"

	// AppDataPath defines path where app data is stored
	AppDataPath = "/var/lib/nordvpn/"

	DatFilesPath = AppDataPath + "data/"

	BakFilesPath = AppDataPath + "backup/"

	// LogFilePath defines CLI log path
	LogFilePath = UserDataPath + "cli.log"

	// OvpnTemplatePath defines filename of ovpn template file
	OvpnTemplatePath = DatFilesPath + "ovpn_template.xslt"

	// OvpnObfsTemplatePath defines filename of ovpn obfuscated template file
	OvpnObfsTemplatePath = DatFilesPath + "ovpn_xor_template.xslt"
)

var (
	PlatformSupportsIPv4 = true
	PlatformSupportsIPv6 = true
)

func GetSupportedIPTables() []string {
	var iptables []string
	if PlatformSupportsIPv4 {
		iptables = append(iptables, "iptables")
	}
	if PlatformSupportsIPv6 {
		iptables = append(iptables, "ip6tables")
	}
	return iptables
}

// GetFilesharedSocket to communicate with fileshare daemon
func GetFilesharedSocket(uid int) string {
	_, err := os.Stat(fmt.Sprintf("/run/user/%d", uid))
	if uid == 0 || os.IsNotExist(err) {
		return fmt.Sprintf("/run/%s/%s.sock", Fileshared, Fileshared)
	}
	return fmt.Sprintf("/run/user/%d/%s/%s.sock", uid, Fileshared, Fileshared)
}

// GetFilesharedConfigDirPath returns the directory used to store nordfileshared logs and transfers history
func GetFilesharedConfigDirPath(homeDirectory string) (string, error) {
	if homeDirectory == "" {
		return "", errors.New("user does not have a home directory")
	}
	// We are running as root, so we cannot retrieve user config directory path dynamically. We
	// hardcode it to /home/<username>/.config, and if it doesn't exist on the expected path
	// (i.e XDG_CONFIG_HOME is set), we default to /var/log/nordvpn/nordfileshared-<username>-<uid>.log
	userConfigPath := filepath.Join(homeDirectory, ConfigDirectory, UserDataPath)
	_, err := os.Stat(userConfigPath)
	if err == nil {
		return userConfigPath, nil
	}

	return "", fmt.Errorf("%s directory not found in users home directory", ConfigDirectory)
}

// GetFilesharedLogPath when logs aren't handled by systemd
func GetFilesharedLogPath(uid string) string {
	filesharedLogFilename := Fileshared + ".log"
	if uid == "0" {
		return filepath.Join(LogPath, filesharedLogFilename)
	}

	usr, err := user.LookupId(uid)
	if err != nil {
		log.Printf("failed to lookup user, users fileshared logs will be stored in %s: %s", LogPath, err.Error())
	}

	configDir, err := GetFilesharedConfigDirPath(usr.HomeDir)

	if err != nil {
		log.Printf("users fileshared logs will be stored in %s: %s", LogPath, err.Error())
		return filepath.Join(LogPath, Fileshared+"-"+uid+".log")
	}

	return filepath.Join(configDir, filesharedLogFilename)
}

// GetNordvpnGid returns id of group defined in NordvpnGroup
func GetNordvpnGid() (int, error) {
	group, err := user.LookupGroup(NordvpnGroup)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(group.Gid)
}

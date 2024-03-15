package domain

import (
	"time"
)

// CustomerConfig - entity of customer, business object
type CustomerConfig struct {
	CustomerID     int64  `json:"customerId"`
	CustomerName   string `json:"customerName"`
	CustomerTypeID int64  `json:"customerTypeId"`
	SdServiceID    int64  `json:"sdServiceId"`
	SdCreatorID    int64  `json:"sdCreatorId"`
	SdDestination  string `json:"sdDestination"`
}

type CustomerConfigs []CustomerConfig

// Project - entity of project, business object
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TypeID      int64  `json:"typeId"`
	TypeName    string `json:"typeName"`
	ParentID    int64  `json:"parentId"`
	ManagerID   int64  `json:"managerId"`
	ManagerName string `json:"managerName"`
	IsEnabled   bool   `json:"isEnabled"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	DBName      string `json:"dbName"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	DBType      int    `json:"db_type"`
}

type Projects []Project

// FTPServer parameters of ftp
type FTPServer struct {
	// Server ip or DNS name of the ftp server
	Server string `json:"server"`
	// User for logit to FTP
	User string `json:"user"`
	// Password for login to FTP
	Password string `json:"password"`
	// Root is subfolder for ftpUser home folder
	Root string `json:"root"`
}

// FTPinfo - information about ftp for project with id
type FTPinfo struct {
	// MainFTP настройки для внешнего взаимодействия с ФТП сервером, в основном из сети Ваткома
	MainFTP FTPServer `json:"mainFtp"`
	// DeviceFTP настройки для устройств выгружающих видео на ФТП
	DeviceFTP FTPServer `json:"deviceFtp"`
	// ProxyFTP настройки для взаимодействия с ФТП сервером неким агентом. ФТП сервер может быть как буферный так и нет
	ProxyFTP FTPServer `json:"proxyFtp"`
	// LocalPathFTP папка на локальном ФТП сервере клиента
	LocalPathFTP string `json:"localPathFtp"`
	// Comment комментарий
	Comment string `json:"comment"`
	// Verified флаг верификации параметров - 1 верифицирвана и параметрам можно доверять, 0 - наоборот
	Verified string `json:"verified"`
	// WhoVerified имя того, кто произвел верификацию
	WhoVerified string `json:"whoVerified"`
	// WhenVerified когда произведена верификация
	WhenVerified string `json:"whenVerified"`
}

// ManualCounting - controller's manual counting value
type ManualCounting struct {
	Chksumin     int       `json:"chkSumin"`
	Fsumin       int       `json:"fSumIn"`
	Sigmasumin   float64   `json:"sigmaSumIn"`
	Chksumout    int       `json:"chkSumOut"`
	Fsumout      int       `json:"fSumOut"`
	Sigmasumout  float64   `json:"sigmaSumOut"`
	Chktimestart time.Time `json:"chkTimeStart"`
	Chktimeend   time.Time `json:"chkTimeEnd"`
	Message      string    `json:"message"`
	Sernomer     string    `json:"sernomer"`
}

type ManualCountings []ManualCounting

// VideocheckConfig propertie of videocheck parameters
type VideocheckConfig struct {
	ProjectID   int64  `json:"projectId"`
	LocalServer bool   `json:"localServer"`
	LocalCam    bool   `json:"localCam"`
	LocalFtp    bool   `json:"localFtp"`
	Options     string `json:"options"`
}

type VideocheckConfigs []VideocheckConfig

type CustomerRepo interface {
	FindCustomersConfig(int64, int64, bool) (CustomerConfigs, int64, error)
	FindCustomerConfig(int64) (*CustomerConfig, error)
	FindProjects(int64, int64, bool) (Projects, int64, error)
	FindProjectByID(int64, *string) (*Project, error)
	FindFTPByID(int64) (*FTPinfo, error)
	FindManualCountings(int64, int64, int64, int64) (ManualCountings, int64, error)
	FindVideoChechCfgs(int64, int64) (VideocheckConfigs, int64, error)
	FindVideoCheckCfgByID(int64) (*VideocheckConfig, error)
	StoreVideoCheckCfg(VideocheckConfig) error
	UpSertVideoCheckCfg(VideocheckConfig) error
	DeleteVideoCheckCfgByID(int64) (int64, error)
	Health() error
}

// [ Intraservice DataBase ]

// Asset - entity of asset, bisness object
type Asset struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	ServiceDeskParentID int64  `json:"serviceDeskParentId"`
	Changed             string `json:"changed"`
	ServiceDeskID       int64  `json:"serviceDeskId"`
}

type Assets []Asset

// AssetRepo vehavior of the Asset repository
type AssetRepo interface {
	FindAll(int64, int64) (Assets, int64, error)
	FindByID(int64) (*Asset, error)
	Health() error
}

// [ Intraservice API ]

// ControllerTask - json string with Task from serviceDesk
type ControllerTask struct {
	Task string `json:"task"`
}

type ControllerTasks []ControllerTask

// TaskStatus - json string with Task Status description from helpdesk
type TaskStatus struct {
	Comment          string `json:"comment"`
	IsPrivateComment bool   `json:"isPrivateComment"`
	StatusID         int    `json:"statusId"`
	ResultFieldName  string `json:"resultFieldName"`
	ResultFieldValue string `json:"resultFieldValue"`
}

// TaskComment - json string with Task Status description from helpdesk
type TaskComment struct {
	Comment string `json:"comment"`
}

// SDRepo behavior of the ServiceDesk repository
type SDRepo interface {
	FindAllBySN(string, int64, int64) (ControllerTasks, int64, error)
	Health() error
	TaskAddComment(string, string) error
	TaskSetStatus(string, TaskStatus) error
}

// [ references ]

// Entity ..
type Entity struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type Entities []Entity

// RefRepo repository od reference data
type RefRepo interface {
	FindEntities(int64, int64) (Entities, int64, error)
	FindEntityByID(string) (*Entity, error)
	Health() error
}

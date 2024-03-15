package repos

import (
	"context"
	"strconv"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/pkg/errors"

	"git.countmax.ru/countmax/cminfo"
)

type CustomersRepo struct {
	connString string
	info       *cminfo.CMINFO
	timeout    time.Duration
}

// NewCustomersRepo returns instance of CustomersRepo with connected CM_INFO DB,
// format connection string: "server=study-app;user id=transport;password=transport;port=1433;database=CM_Transport523;"
func NewCustomersRepo(connString string, timeout time.Duration) (*CustomersRepo, error) {
	cr := &CustomersRepo{
		connString: connString,
		timeout:    timeout,
	}
	info, err := cminfo.New(connString)
	if err != nil {
		return nil, err
	}
	cr.info = info
	return cr, nil
}

// FindCustomersConfig implementation of CustomrRepo interface,
// returns slice of domain.CustomerConfig
func (cr *CustomersRepo) FindCustomersConfig(offset int64, limit int64,
	enabled bool) (domain.CustomerConfigs, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoCustomers, count, err := cr.info.GetObjectConfigsWithContext(ctx, offset, limit, enabled)
	if err != nil {
		return nil, count, err
	}
	result := make(domain.CustomerConfigs, len(infoCustomers))
	for i, v := range infoCustomers {
		result[i].CustomerID = v.CustomerID
		result[i].CustomerName = v.CustomerName
		result[i].CustomerTypeID = v.CustomerTypeID
		result[i].SdServiceID = v.SdServiceID
		result[i].SdCreatorID = v.SdCreatorID
		result[i].SdDestination = v.SdDestination
	}
	return result, count, nil
}

// FindCustomerConfig implementation of CustomrRepo interface,
// returns domain.CustomerConfig
func (cr *CustomersRepo) FindCustomerConfig(id int64) (*domain.CustomerConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoCustomer, err := cr.info.GetObjectConfigByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	customer := &domain.CustomerConfig{
		CustomerID:     infoCustomer.CustomerID,
		CustomerName:   infoCustomer.CustomerName,
		CustomerTypeID: infoCustomer.CustomerTypeID,
		SdServiceID:    infoCustomer.SdServiceID,
		SdCreatorID:    infoCustomer.SdCreatorID,
		SdDestination:  infoCustomer.SdDestination,
	}
	return customer, nil
}

// FindProjects implementation of CustomrRepo interface,
// returns slice of domain.Project
func (cr *CustomersRepo) FindProjects(offset, limit int64, enabled bool) (domain.Projects, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoCustomers, count, err := cr.info.GetObjectConfigsDBWithContext(ctx, offset, limit, enabled)
	if err != nil {
		return nil, count, err
	}
	result := make(domain.Projects, len(infoCustomers))
	for i, v := range infoCustomers {
		result[i].ID = v.ID
		result[i].Name = v.Name
		result[i].TypeID = v.TypeID
		result[i].TypeName = v.TypeName
		result[i].ParentID = v.ParentID
		result[i].ManagerID = v.ManagerID
		result[i].ManagerName = v.ManagerName
		result[i].IsEnabled = v.IsEnabled
		result[i].IP = v.IP
		result[i].Port = v.Port
		result[i].DBName = v.DBName
		result[i].Login = v.Login
		result[i].Password = v.Password
		result[i].DBType = v.DBType
	}
	return result, count, nil
}

// FindProjectByID implementation of CustomrRepo interface,
// returns domain.Project with specified Id
func (cr *CustomersRepo) FindProjectByID(id int64, dbType *string) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	if dbType == nil {
		infoCustomer, err := cr.info.GetObjectConfigDBWithContext(ctx, id)
		if err != nil {
			return nil, errors.WithMessage(err, "info.GetObjectConfigDBWithContext error")
		}
		return infoProject2domainProject(infoCustomer), nil
	}
	//
	dbTypeInt, err := strconv.Atoi(*dbType)
	if err != nil {
		return nil, errors.WithMessage(err, "parse to int")
	}

	infoCustomer, err := cr.info.GetObjectConfigDBType(ctx, id, dbTypeInt)
	if err != nil {
		return nil, errors.WithMessage(err, "info.GetObjectConfigDBType error")
	}
	return infoProject2domainProject(infoCustomer), nil
}

// helper

func infoProject2domainProject(src *cminfo.Project) *domain.Project {
	if src == nil {
		return nil
	}
	return &domain.Project{
		ID:          src.ID,
		Name:        src.Name,
		TypeID:      src.TypeID,
		TypeName:    src.TypeName,
		ParentID:    src.ParentID,
		ManagerID:   src.ManagerID,
		ManagerName: src.ManagerName,
		IsEnabled:   src.IsEnabled,
		IP:          src.IP,
		Port:        src.Port,
		DBName:      src.DBName,
		Login:       src.Login,
		Password:    src.Password,
		DBType:      src.DBType,
	}
}

// FindFTPByID implementation of CustomrRepo interface,
// returns FTP settings for customer
func (cr *CustomersRepo) FindFTPByID(id int64) (*domain.FTPinfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoFTP, err := cr.info.GetObjectFTPContext(ctx, id)
	if err != nil {
		return nil, err
	}
	ftp := &domain.FTPinfo{
		MainFTP: domain.FTPServer{
			Server:   infoFTP.MainFTP.Server,
			User:     infoFTP.MainFTP.User,
			Password: infoFTP.MainFTP.Password,
			Root:     infoFTP.MainFTP.Root,
		},
		DeviceFTP: domain.FTPServer{
			Server:   infoFTP.DeviceFTP.Server,
			User:     infoFTP.DeviceFTP.User,
			Password: infoFTP.DeviceFTP.Password,
			Root:     infoFTP.DeviceFTP.Root,
		},
		ProxyFTP: domain.FTPServer{
			Server:   infoFTP.ProxyFTP.Server,
			User:     infoFTP.ProxyFTP.User,
			Password: infoFTP.ProxyFTP.Password,
			Root:     infoFTP.ProxyFTP.Root,
		},
		LocalPathFTP: infoFTP.LocalPathFTP,
		Comment:      infoFTP.Comment,
		Verified:     infoFTP.Verified,
		WhoVerified:  infoFTP.WhoVerified,
		WhenVerified: infoFTP.WhenVerified,
	}
	return ftp, nil
}

// FindManualCountings implementation of CustomrRepo interface,
// returns manual countings data for controller id in the project of customer pid
func (cr *CustomersRepo) FindManualCountings(id, cid, offset, limit int64) (domain.ManualCountings, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoMCs, count, err := cr.info.GetManualCntsByCntrlIDContext(ctx, id, cid, offset, limit)
	if err != nil {
		return nil, count, err
	}
	result := make(domain.ManualCountings, len(infoMCs))
	for i, v := range infoMCs {
		result[i].Chksumin = v.Chksumin
		result[i].Fsumin = v.Fsumin
		result[i].Sigmasumin = v.Sigmasumin
		result[i].Chksumout = v.Chksumout
		result[i].Fsumout = v.Fsumout
		result[i].Sigmasumout = v.Sigmasumout
		result[i].Chktimestart = v.Chktimestart
		result[i].Chktimeend = v.Chktimeend
		result[i].Message = v.Message
		result[i].Sernomer = v.Sernomer
	}
	return result, count, nil
}
func (cr *CustomersRepo) FindVideoChechCfgs(offset, limit int64) (domain.VideocheckConfigs, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoVCs, count, err := cr.info.GetVideoCheckContext(ctx, offset, limit)
	if err != nil {
		return nil, count, err
	}
	result := make(domain.VideocheckConfigs, len(infoVCs))
	for i, v := range infoVCs {
		result[i].ProjectID = v.ProjectID
		result[i].LocalServer = v.LocalServer
		result[i].LocalCam = v.LocalCam
		result[i].LocalFtp = v.LocalFtp
		result[i].Options = v.Options
	}
	return result, count, nil
}

func (cr *CustomersRepo) FindVideoCheckCfgByID(id int64) (*domain.VideocheckConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	infoVC, err := cr.info.GetVideoCheckByIDContext(ctx, id)
	if err != nil {
		return nil, err
	}
	if infoVC == nil {
		return nil, nil
	}
	vc := &domain.VideocheckConfig{
		ProjectID:   infoVC.ProjectID,
		LocalServer: infoVC.LocalServer,
		LocalCam:    infoVC.LocalCam,
		LocalFtp:    infoVC.LocalFtp,
		Options:     infoVC.Options,
	}
	return vc, nil
}

func (cr *CustomersRepo) StoreVideoCheckCfg(newCfg domain.VideocheckConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	newInfoVC := cminfo.VideocheckConfig{
		ProjectID:   newCfg.ProjectID,
		LocalServer: newCfg.LocalServer,
		LocalCam:    newCfg.LocalCam,
		LocalFtp:    newCfg.LocalFtp,
		Options:     newCfg.Options,
	}
	return cr.info.NewVideoCheckConfigContext(ctx, newInfoVC)
}

func (cr *CustomersRepo) UpSertVideoCheckCfg(updCfg domain.VideocheckConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	updInfoVC := cminfo.VideocheckConfig{
		ProjectID:   updCfg.ProjectID,
		LocalServer: updCfg.LocalServer,
		LocalCam:    updCfg.LocalCam,
		LocalFtp:    updCfg.LocalFtp,
		Options:     updCfg.Options,
	}
	return cr.info.UpdVideoCheckConfigContext(ctx, updInfoVC)
}

func (cr *CustomersRepo) DeleteVideoCheckCfgByID(id int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	return cr.info.DelVideoCheckConfigContext(ctx, id)
}

// Health implementation of CustomrRepo interface,
// check simple sql query to sql server
func (cr *CustomersRepo) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), cr.timeout)
	defer cancel()
	return cr.info.Healthz(ctx)
}

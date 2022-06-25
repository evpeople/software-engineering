package constants

const (
	MySQLDefaultDSN = "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local"
	UserTableName   = "user"
	CarsTableName   = "cars"
	BillTableName   = "bills"
	PileTableName   = "piles"
	SecretKey       = "secret key"
	TimeLayoutStr   = "2006-01-02 15:04:05"
	QuickCharge     = 0
	Scale           = 10 // 比例尺，测试1min=实际时间10min
)

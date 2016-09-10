package influxdbhelper

import (
	"github.com/vlad-doru/influxus"
	"github.com/influxdata/influxdb/client/v2"
	uct "uct/common"
	log "github.com/Sirupsen/logrus"
)

var (
	influxClient client.Client
	AuditLogger *log.Logger
)

func initInflux(config *uct.Config) {
	var err error
	// Create the InfluxDB client.
	influxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.GetInfluxAddr(),
		Password: config.InfluxDb.Password,
		Username: config.InfluxDb.User,
	})

	if err != nil {
		log.Fatalf("Error while creating the client: %v", err)
	}

	// Create and add the hook.
	auditHook, err := influxus.NewHook(
		&influxus.Config{
			Client:             influxClient,
			Database:           "universityct", // DATABASE MUST BE CREATED
			DefaultMeasurement: "hermes_ops",
			BatchSize:          1, // default is 100
			BatchInterval:      1, // default is 5 seconds
			Tags:               []string{"university_name"},
			Precision: "ms",
		})

	uct.CheckError(err)

	// Add the hook to the standard logger.
	AuditLogger = log.New()
	AuditLogger.Hooks.Add(auditHook)
}

func GetClient(config uct.Config) (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.GetInfluxAddr(),
		Password: config.InfluxDb.Password,
		Username: config.InfluxDb.User,
		UserAgent: config.AppName,
	})
}
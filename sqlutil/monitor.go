package sqlutil

import (
	"context"
	"database/sql"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/stats"
)

// MonitorDatabase collects database connection pool metrics at regular intervals
func MonitorDatabase(
	ctx context.Context,
	conf *config.Config,
	statsFactory stats.Stats,
	db *sql.DB,
	dbIdentifier string,
) {
	reportInterval := conf.GetDurationVar(10, time.Second, "Database.ReportInterval", dbIdentifier+".Database.ReportInterval")

	maxOpenConnectionsStat := statsFactory.NewStat(dbIdentifier+".db.max_open_connections", stats.CountType)
	openConnectionsStat := statsFactory.NewStat(dbIdentifier+".db.open_connections", stats.CountType)
	inUseStat := statsFactory.NewStat(dbIdentifier+".db.in_use", stats.CountType)
	idleStat := statsFactory.NewStat(dbIdentifier+".db.idle", stats.CountType)
	waitCountStat := statsFactory.NewStat(dbIdentifier+".db.wait_count", stats.CountType)
	waitDurationStat := statsFactory.NewStat(dbIdentifier+".db.wait_duration", stats.TimerType)
	maxIdleClosedStat := statsFactory.NewStat(dbIdentifier+".db.max_idle_closed", stats.CountType)
	maxIdleTimeClosedStat := statsFactory.NewStat(dbIdentifier+".db.max_idle_time_closed", stats.CountType)
	maxLifetimeClosedStat := statsFactory.NewStat(dbIdentifier+".db.max_lifetime_closed", stats.CountType)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(reportInterval):
				dbStats := db.Stats()

				maxOpenConnectionsStat.Count(dbStats.MaxOpenConnections)
				openConnectionsStat.Count(dbStats.OpenConnections)
				inUseStat.Count(dbStats.InUse)
				idleStat.Count(dbStats.Idle)
				waitCountStat.Count(int(dbStats.WaitCount))
				waitDurationStat.SendTiming(dbStats.WaitDuration)
				maxIdleClosedStat.Count(int(dbStats.MaxIdleClosed))
				maxIdleTimeClosedStat.Count(int(dbStats.MaxIdleTimeClosed))
				maxLifetimeClosedStat.Count(int(dbStats.MaxLifetimeClosed))
			}
		}
	}()
}

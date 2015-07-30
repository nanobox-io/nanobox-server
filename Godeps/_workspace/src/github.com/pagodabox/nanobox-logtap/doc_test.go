package logtap_test

import "github.com/nanobox-core/logtap"
import "github.com/jcelliott/lumber"
import "time"

func ExampleLogtap() {
	ltap := logtap.New(logger)
	ltap.Start()

	// fire up a collector
	sysc := logtap.NewSyslogCollector(514)
	ltap.AddCollector("syslog", sysc)
	sysc.Start()

	// start draining to concole
	conc := logtap.NewConsoleDrain()
	ltap.AddDrain("concole", conc)

	// start a historical drain
	hist := logtap.NewHistoricalDrain(8080, "./bolt.db", 1000)
	hist.Start()
	ltap.AddDrain("history", hist)

	time.Sleep(1000 * time.Second)
}

func ExampleConcoleDrain() {
	// no params are necessary
	// really simple
	conc := logtap.NewConsoleDrain()
}

func ExampleHistoricalDrain() {
	// start a historical drain
	hist := logtap.NewHistoricalDrain(8080, "./bolt.db", 1000)
	hist.Start()
	ltap.AddDrain("history", hist)

}

func ExampleSyslogCollector() {
	ltap := logtap.New(nil)

	// fire up a collector
	sysc := logtap.NewSyslogCollector(514)
	ltap.AddCollector("syslog", sysc)
	sysc.Start()

}

func ExamplePublishDrain() {
	ltap := logtap.New(nil)

	pub := logtap.newPublishDrain(publisher)
	ltap.AddDrain("publisher", pub)
}

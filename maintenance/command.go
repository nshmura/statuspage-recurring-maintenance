package maintenance

import (
	"flag"
	"log"
	"os"
	"time"
)

type Command interface {
	Run()
}

var dateLayout = "2006-01-02"

var loc, _ = time.LoadLocation("Asia/Tokyo")

// コマンドライン引数から、実行するコマンド情報を読み込む
func ReadCommand() Command {
	accessToken := os.Getenv("STATUSPAGE_ACCESS_KEY")

	recurringCmd := flag.NewFlagSet("recurring", flag.ExitOnError)
	recurringScheduleFilename := recurringCmd.String("schedule", "", "file to load maintenance schedule information")
	recurringFrom := recurringCmd.String("from", "", "first date to create schedule")
	recurringDay := recurringCmd.Int("day", 0, "days of terms to create schedule")
	recurringStatuspageFilename := recurringCmd.String("statuspage", "", "file to load configuration of statuspage")
	recurringDryRun := recurringCmd.Bool("dryRun", false, "is dryRun")

	flag.Parse()

	switch os.Args[1] {

	case "recurring":
		recurringCmd.Parse(os.Args[2:])
		fromDate, err := time.ParseInLocation(dateLayout, *recurringFrom, loc)
		if err != nil {
			log.Fatalf("[ERROR] invalid fromDate: %s", err)
		}

		return &RecurringCommand{
			isDryRun:           *recurringDryRun,
			ScheduleFilename:   *recurringScheduleFilename,
			StatuspageFilename: *recurringStatuspageFilename,
			FromDate:           fromDate,
			ToDate:             fromDate.AddDate(0, 0, *recurringDay-1),
			AccessToken:        accessToken,
		}

	default:
		log.Fatalf("[ERROR] Unknown command is specified")
		return nil
	}
}

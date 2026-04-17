package quota

import (
	"errors"
	"fmt"
	"strings"

	"github.com/floj/gotimekpr/pkg/db"
)

type QuotaManager struct {
	dbq *db.Queries
}

func WeekdaysFromStrings(ss []string) ([]int64, error) {
	wds := []int64{}
	errs := []error{}

	for _, s := range ss {
		s = strings.ToLower(s)
		switch strings.ToLower(s) {
		case "all", "*":
			return []int64{0, 1, 2, 3, 4, 5, 6}, nil
		case "weekend":
			wds = append(wds, 0, 6)
			continue
		case "workdays":
			wds = append(wds, 1, 2, 3, 4, 5)
			continue
		default:
			wd, err := WeekdayFromString(s)
			errs = append(errs, err)
			wds = append(wds, wd)
		}
	}
	return wds, errors.Join(errs...)
}

func WeekdayFromString(s string) (int64, error) {
	switch strings.ToLower(s) {
	case "sunday", "sun", "su", "0":
		return 0, nil
	case "monday", "mon", "mo", "1":
		return 1, nil
	case "tuesday", "tue", "tu", "2":
		return 2, nil
	case "wednesday", "wed", "we", "3":
		return 3, nil
	case "thursday", "thu", "th", "4":
		return 4, nil
	case "friday", "fri", "fr", "5":
		return 5, nil
	case "saturday", "sat", "sa", "6":
		return 6, nil
	default:
		return -1, fmt.Errorf("invalid weekday: %s", s)
	}
}

func WeekdayToString(wd int64) string {
	switch wd {
	case 0:
		return "Sunday"
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	default:
		return fmt.Sprintf("Invalid weekday: %d", wd)
	}
}

func NewQuotaManager(dbq *db.Queries) *QuotaManager {
	return &QuotaManager{
		dbq: dbq,
	}
}

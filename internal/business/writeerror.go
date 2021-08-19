package business

import (
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func Rewrite(new, last payment.Payment, lastWriteError uint64) bool {
	if last == nil || new.UID() != lastWriteError || new.UID() != last.UID() {
		return false
	}
	// logs.LogInfo.Printf("new: %v, last: %v",
	// 	new.Historical()[len(new.Historical())-1].TimeTransaction(),
	// 	last.Historical()[len(last.Historical())-2].TimeTransaction())
	if len(new.Historical()) < 1 || len(last.Historical()) < 2 {
		return false
	}
	if new.Historical()[len(new.Historical())-1].TimeTransaction().
		After(last.Historical()[len(last.Historical())-2].TimeTransaction()) {
		return false
	}

	return true
}

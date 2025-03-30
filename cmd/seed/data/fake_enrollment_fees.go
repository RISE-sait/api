package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

func GetEnrollmentFees(programIDs, membershipIDs []uuid.UUID) dbSeed.InsertEnrollmentFeesParams {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		programIDArray    []uuid.UUID
		membershipIDArray []uuid.UUID
		dropInPriceArray  []decimal.Decimal
		programPriceArray []decimal.Decimal
	)

	for _, programID := range programIDs {
		// First add the non-member pricing (membership_id = nil)
		programIDArray = append(programIDArray, programID)
		membershipIDArray = append(membershipIDArray, uuid.Nil)                                 // Nil UUID for non-member
		dropInPriceArray = append(dropInPriceArray, decimal.NewFromFloat(r.Float64()*100+10))   // $10-$110
		programPriceArray = append(programPriceArray, decimal.NewFromFloat(r.Float64()*500+50)) // $50-$550

		// Member pricing
		for _, membershipID := range membershipIDs {
			programIDArray = append(programIDArray, programID)
			membershipIDArray = append(membershipIDArray, membershipID)

			if rand.Float32() < 0.7 { // 70% eligible
				// Set both prices for eligible members
				dropInPriceArray = append(dropInPriceArray, decimal.NewFromFloat(r.Float64()*80+5))     // $5-$85
				programPriceArray = append(programPriceArray, decimal.NewFromFloat(r.Float64()*400+40)) // $40-$440
			} else { // 30% not eligible
				// Set program price to 9999 and drop-in to 0 (will be NULL)
				dropInPriceArray = append(dropInPriceArray, decimal.NewFromInt(0))
				programPriceArray = append(programPriceArray, decimal.NewFromInt(9999))
			}
		}
	}

	return dbSeed.InsertEnrollmentFeesParams{
		ProgramIDArray:    programIDArray,
		MembershipIDArray: membershipIDArray,
		DropInPriceArray:  dropInPriceArray,
		ProgramPriceArray: programPriceArray,
	}
}

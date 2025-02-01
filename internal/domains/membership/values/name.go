package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type MembershipName struct {
	value string
}

func NewMembershipName(name string) (*MembershipName, *errLib.CommonError) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errLib.New("'name' cannot be empty or whitespace", http.StatusBadRequest)
	}
	return &MembershipName{value: name}, nil
}

func (c MembershipName) String() string {
	return c.value
}

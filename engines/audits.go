package engines

import (
	"net/http"
	"time"
)

// AUDIT_DATE_LAYOUT defines the date format
const AUDIT_DATE_LAYOUT = "20060102"

// EndpointRootAuditLogs displays audit logs (no possibility to change them)
func EndpointRootAuditLogs(c *HandlerContext) error {
	parameters := c.RequestUrlParameters()
	var from, to string
	for k, v := range parameters {
		switch k {
		case "from":
			if len(v) != 1 || !ValidateDateFormat(v[0]) {
				c.Build(http.StatusBadRequest, "invalid parameter from: expecting one value as a eight digit number", nil)
				return nil
			} else {
				from = v[0]
			}
		case "to":
			if len(v) != 1 || !ValidateDateFormat(v[0]) {
				c.Build(http.StatusBadRequest, "invalid parameter to: expecting one value as a eight digit number", nil)
				return nil
			} else {
				to = v[0]
			}
		default:
			c.Build(http.StatusBadRequest, "invalid parameter. Expecting only from and to", nil)
			return nil
		}
	}

	// from and to parameters found, if any
	if from != "" && to != "" && from > to {
		c.Build(http.StatusBadRequest, "invalid parameters: from is greater than to", nil)
		return nil
	}

	// parse them to become a date
	var fromDate, toDate time.Time
	if len(from) == len(AUDIT_DATE_LAYOUT) {
		if fd, err := time.Parse(AUDIT_DATE_LAYOUT, from); err != nil {
			c.Build(http.StatusBadRequest, "invalid from parameter: format mismatch", nil)
			return nil
		} else {
			fromDate = fd
		}
	}

	if len(to) == len(AUDIT_DATE_LAYOUT) {
		if fd, err := time.Parse(AUDIT_DATE_LAYOUT, to); err != nil {
			c.Build(http.StatusBadRequest, "invalid to parameter: format mismatch", nil)
			return nil
		} else {
			toDate = fd
		}
	} else {
		toDate = time.Now()
	}

	// get answer and returns it
	if values, err := c.Dao.LoadAuditEvents(c.GetCurrentContext(), fromDate, toDate); err != nil {
		c.Build(http.StatusInternalServerError, "invalid to parameter: format mismatch", nil)
		return nil
	} else {
		c.BuildJson(http.StatusOK, values, c.RequestHeaderByNames("Authorization"))
		return nil
	}
}

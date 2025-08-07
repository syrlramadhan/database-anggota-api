package util

import (
    "database/sql/driver"
    "fmt"
    "time"
)

type CustomDate struct {
    time.Time
}

// Format tampilan JSON (optional, tergantung kebutuhan)
const layout = "2006-01-02"

// MarshalJSON: mengubah ke format string saat dikirim via JSON
func (cd CustomDate) MarshalJSON() ([]byte, error) {
    return []byte(`"` + cd.Format(layout) + `"`), nil
}

// UnmarshalJSON: mengubah dari string ke time.Time saat parsing JSON
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
    parsedTime, err := time.Parse(`"`+layout+`"`, string(b))
    if err != nil {
        return err
    }
    cd.Time = parsedTime
    return nil
}

// Implement driver.Valuer untuk database insert/update
func (cd CustomDate) Value() (driver.Value, error) {
    return cd.Time, nil
}

// Implement sql.Scanner untuk membaca dari database
func (cd *CustomDate) Scan(value interface{}) error {
    switch v := value.(type) {
    case time.Time:
        cd.Time = v
        return nil
    case []byte:
        t, err := time.Parse(layout, string(v))
        if err != nil {
            return err
        }
        cd.Time = t
        return nil
    case string:
        t, err := time.Parse(layout, v)
        if err != nil {
            return err
        }
        cd.Time = t
        return nil
    }
    return fmt.Errorf("unsupported Scan type for CustomDate: %T", value)
}

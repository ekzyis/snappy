package sn

import (
	"fmt"
	"net/http"
)

func RefreshSession() error {
	req, err := http.NewRequest("GET", SnUrl+"/api/auth/session", nil)
	if err != nil {
		err = fmt.Errorf("error preparing SN request: %w", err)
		return err
	}
	req.Header.Set("Cookie", SnAuthCookie)

	client := http.DefaultClient
	_, err = client.Do(req)
	if err != nil {
		err = fmt.Errorf("error refreshing SN session: %w", err)
		return err
	}
	return nil
}

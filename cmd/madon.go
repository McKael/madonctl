// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"strconv"
	"strings"

	"github.com/McKael/madon/v3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var scopes = []string{"read", "write", "follow"}

func madonInit(signIn bool) error {
	if gClient == nil {
		if err := madonInitClient(); err != nil {
			return err
		}
	}
	if signIn {
		return madonLogin()
	}
	return nil
}

func madonInitClient() error {
	if gClient != nil {
		return nil
	}
	var err error

	// Overwrite variables using Viper
	instanceURL = viper.GetString("instance")
	appID = viper.GetString("app_id")
	appSecret = viper.GetString("app_secret")

	if instanceURL == "" {
		return errors.New("no instance provided")
	}

	if verbose {
		errPrint("Instance: '%s'", instanceURL)
	}

	if appID != "" && appSecret != "" {
		// We already have an app key/secret pair
		gClient, err = madon.RestoreApp(AppName, instanceURL, appID, appSecret, nil)
		if err != nil {
			return err
		}
		// Check instance
		if _, err := gClient.GetCurrentInstance(); err != nil {
			return errors.Wrap(err, "could not connect to server with provided app ID/secret")
		}
		if verbose {
			errPrint("Using provided app ID/secret")
		}
		return nil
	}

	if appID != "" || appSecret != "" {
		errPrint("Warning: provided app id/secrets incomplete -- registering again")
	}

	gClient, err = madon.NewApp(AppName, AppWebsite, scopes, madon.NoRedirect, instanceURL)
	if err != nil {
		return errors.Wrap(err, "app registration failed")
	}

	errPrint("Registered new application.")
	return nil
}

func madonLogin() error {
	if gClient == nil {
		return errors.New("application not registered")
	}

	token = viper.GetString("token")
	login = viper.GetString("login")
	password = viper.GetString("password")

	if token != "" { // TODO check token validity?
		if verbose {
			errPrint("Reusing existing token.")
		}
		gClient.SetUserToken(token, login, password, []string{})
		return nil
	}

	err := gClient.LoginBasic(login, password, scopes)
	if err == nil {
		return nil
	}
	if !verbose && err.Error() == "cannot unmarshal server response: invalid character '<' looking for beginning of value" {
		return errors.New("login failed (server did not return a JSON response - check your credentials)")
	}
	return errors.Wrap(err, "login failed")
}

// splitIDs splits a list of IDs into an int64 array
func splitIDs(ids string) (list []int64, err error) {
	var i int64
	if ids == "" {
		return
	}
	l := strings.Split(ids, ",")
	for _, s := range l {
		i, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return
		}
		list = append(list, i)
	}
	return
}

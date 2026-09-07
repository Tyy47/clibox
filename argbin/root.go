package argbin

import (
	"errors"
	"fmt"
	"log"
)

// const array of a collection of root related errors
var (
	ErrNilRoot = errors.New("root cannot be nil")
	ErrEmptyRootName = errors.New("appname cannot be blank")
	ErrEmptyVersionNumber = errors.New("version number cannot be blank")
)


type Root struct {
	// AppName stores the name of your application.
	AppName string
	
	// AppVersion stores the version of your application.
	AppVersion string 
}

// GetAppName returns the name of the application.
func (r *Root) GetAppName() (string, error) {
	if r == nil {
		return "", ErrNilRoot
	}
	if r.AppName == "" {
		err := fmt.Errorf("%w: setting name to default 'appname'.", ErrEmptyRootName)
		log.Println(err)
		r.SetAppName("appname")
	}
	return r.AppName, nil
}

// SetAppName sets the name of the application.
func (r *Root) SetAppName(name string) error {
	if name == "" {
		err := fmt.Errorf("%w: setting name to default 'appname'.", ErrEmptyRootName)
		log.Println(err)
		r.SetAppName("appname")
		return nil
	}
	r.AppName = name
	return nil
}

// GetAppVersion returns the version of the application.
func (r *Root) GetAppVersion() (string, error) {
	if r == nil {
		return "", ErrNilRoot
	}
	if r.AppVersion == "" {
		err := fmt.Errorf("%w: setting version to default '1.0.0'.", ErrEmptyVersionNumber)
		log.Println(err)
		r.SetAppVersion("1.0.0")
	}
	return r.AppVersion, nil
}

// SetAppVersion sets the version of the application.
func (r *Root) SetAppVersion(version string) error {
	if version == "" {
		err := fmt.Errorf("%w: setting version to default '1.0.0'.", ErrEmptyVersionNumber)
		log.Println(err)
		r.SetAppVersion("1.0.0")
		return nil
	}
	r.AppVersion = version
	return nil
}

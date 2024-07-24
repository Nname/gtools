package ldap

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/util/gconv"
)

type AppService struct {
	Addr       string
	Bind       string
	Auth       string
	BaseDn     string
	Filter     string
	Attributes []string
}

func (s *AppService) Conn() (*ldap.Conn, error) {
	dial, err := ldap.Dial("tcp", s.Addr)
	if err != nil {
		return nil, err
	}
	err = dial.Bind(s.Bind, s.Auth)
	if err != nil {
		return nil, err
	}
	return dial, nil
}

func (s *AppService) Get(uid string) (map[string]string, error) {
	conn, err := s.Conn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	searchFilter := strings.ReplaceAll(s.Filter, "*", uid)
	sql := ldap.NewSearchRequest(s.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter, s.Attributes,
		nil,
	)
	requestResult, requestErr := conn.Search(sql)
	if requestErr != nil {
		return nil, requestErr
	}
	if len(requestResult.Entries) > 1 || len(requestResult.Entries) == 0 {
		return nil, errors.New("the results are too many or nonexistent")
	}

	Map := make(map[string]string, 0)
	for index := 0; index <= len(s.Attributes)-1; index++ {
		Map[s.Attributes[index]] = requestResult.Entries[0].GetAttributeValue(s.Attributes[index])
	}
	Map["DN"] = requestResult.Entries[0].DN
	return Map, nil
}

func (s *AppService) Search() ([]map[string]string, error) {
	conn, err := s.Conn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := ldap.NewSearchRequest(s.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		s.Filter, s.Attributes,
		nil,
	)
	requestResult, requestErr := conn.Search(sql)
	if requestErr != nil {
		return nil, requestErr
	}
	var resultList []map[string]string
	for item := 0; item <= len(requestResult.Entries)-1; item++ {
		Map := make(map[string]string, 0)
		for index := 0; index <= len(s.Attributes)-1; index++ {
			Map[s.Attributes[index]] = requestResult.Entries[item].GetAttributeValue(s.Attributes[index])
		}
		Map["DN"] = requestResult.Entries[item].DN
		resultList = append(resultList, Map)
	}
	return resultList, nil
}

func (s *AppService) SearchPage() ([]map[string]string, error) {
	conn, err := s.Conn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := ldap.NewSearchRequest(s.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		s.Filter, s.Attributes,
		nil,
	)
	requestResult, requestErr := conn.SearchWithPaging(sql, 100)
	if requestErr != nil {
		return nil, requestErr
	}
	var resultList []map[string]string
	for item := 0; item <= len(requestResult.Entries)-1; item++ {
		Map := make(map[string]string, 0)
		for index := 0; index <= len(s.Attributes)-1; index++ {
			Map[s.Attributes[index]] = requestResult.Entries[item].GetAttributeValue(s.Attributes[index])
		}
		Map["DN"] = requestResult.Entries[item].DN
		resultList = append(resultList, Map)
	}
	return resultList, nil
}

func (s *AppService) Authentication(username, password string) bool {
	conn, err := s.Conn()
	if err != nil {
		fmt.Println("err, ", err)
		return false
	}
	defer conn.Close()
	getDn, getErr := s.Get(username)
	fmt.Println("getDn, ", getDn)
	if getErr != nil {
		return false
	}
	DnString, _ := getDn["DN"]
	err = conn.Bind(DnString, password)
	if err != nil {
		return false
	}
	return true
}

func (s *AppService) GetMaxId() (int, error) {
	searchData, err := s.SearchPage()
	if err != nil {
		return 0, err
	}
	var uidList garray.IntArray
	for _, item := range searchData {
		for key, value := range item {
			if key == "uidNumber" {
				uidList.Append(gconv.Int(value))
			}
		}
	}
	return uidList.Sort(true).At(0), nil
}

func (s *AppService) UpdatePwd(dn, passwd string) error {
	conn, err := s.Conn()
	if err != nil {
		return err
	}
	setPassRequest := ldap.NewPasswordModifyRequest(dn, "", passwd)
	_, err = conn.PasswordModify(setPassRequest)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppService) AddDn(dn, pwd string, attrs map[string][]string) (bool, error) {
	conn, err := s.Conn()
	if err != nil {
		return false, err
	}
	defer conn.Close()
	addRequest := ldap.NewAddRequest(dn, nil)
	for key, value := range attrs {
		addRequest.Attribute(key, value)
	}
	if conn.Add(addRequest) != nil {
		return false, err
	}
	if pwd != "" {
		if s.UpdatePwd(dn, pwd) != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *AppService) DeleteDn(dn string) error {
	conn, err := s.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	delRequest := ldap.NewDelRequest(dn, nil)
	err = conn.Del(delRequest)
	if err != nil {
		return err
	}
	return nil
}

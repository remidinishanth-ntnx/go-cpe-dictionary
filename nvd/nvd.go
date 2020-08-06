package nvd

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/knqyf263/go-cpe/common"
	"github.com/knqyf263/go-cpe/naming"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	c "github.com/remidinishanth/go-cpe-dictionary/config"
	"github.com/remidinishanth/go-cpe-dictionary/db"
	"github.com/remidinishanth/go-cpe-dictionary/models"
)

// List has cpe-item list
// https://nvd.nist.gov/cpe.cfm
type List struct {
	Items []Item `xml:"cpe-item"`
}

// Item has CPE information
type Item struct {
	Name      string    `xml:"name,attr"`
	Cpe23Item Cpe23Item `xml:"cpe23-item"`
	Titles    []Title   `xml:"title"`

	// each items
	//  Part     string
	//  Vendor   string
	//  Product  string
	//  Version  string
	//  Update   string
	//  Edition  string
	//  Language string
}

// Cpe23Item : Cpe23Item
type Cpe23Item struct {
	Name string `xml:"name,attr"`
}

// Title has title, lang
type Title struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

// FetchAndInsertCPE : FetchAndInsertCPE
func FetchAndInsertCPE(driver db.DB) (err error) {
	if err = FetchAndInsertCpeDictioanry(driver); err != nil {
		return fmt.Errorf("Failed to fetch cpe dictionary. err : %s", err)
	}
	return nil
}

// FetchAndInsertCpeDictioanry : FetchCPE
func FetchAndInsertCpeDictioanry(driver db.DB) (err error) {
	var cpeDictionary List
	var body string
	var errs []error
	var resp *http.Response
	url := "http://nvd.nist.gov/feeds/xml/cpe/dictionary/official-cpe-dictionary_v2.3.xml.gz"
	resp, body, errs = gorequest.New().Proxy(c.Conf.HTTPProxy).Get(url).End()
	if len(errs) > 0 || resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error. errs: %v, url: %s", errs, url)
	}

	b := bytes.NewBufferString(body)
	reader, err := gzip.NewReader(b)
	defer func() {
		_ = reader.Close()
	}()
	if err != nil {
		return fmt.Errorf("Failed to decompress NVD feedfile. url: %s, err: %s", url, err)
	}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to Read NVD feedfile. url: %s, err: %s", url, err)
	}
	if err = xml.Unmarshal(bytes, &cpeDictionary); err != nil {
		return fmt.Errorf("Failed to unmarshal. url: %s, err: %s", url, err)
	}

	var cpes []*models.CategorizedCpe
	if cpes, err = ConvertNvdCpeDictionaryToModel(cpeDictionary); err != nil {
		return err
	}

	if err = driver.InsertCpes(cpes); err != nil {
		return fmt.Errorf("Failed to insert cpes. err : %s", err)
	}
	return nil
}

// ConvertNvdCpeDictionaryToModel :
func ConvertNvdCpeDictionaryToModel(nvd List) (cpes []*models.CategorizedCpe, err error) {
	for _, item := range nvd.Items {
		var wfn common.WellFormedName
		if wfn, err = naming.UnbindFS(item.Cpe23Item.Name); err != nil {
			return nil, errors.Wrapf(err, "Failed to unbind cpe fs: %s", item.Cpe23Item.Name)
		}
		for _, title := range item.Titles {
			if title.Lang == "en-US" {
				cpes = append(cpes, &models.CategorizedCpe{
					Title:  title.Value,
					CpeURI: naming.BindToURI(wfn),
					CpeFS:  naming.BindToFS(wfn),
				})
			}
		}
	}
	return cpes, nil
}

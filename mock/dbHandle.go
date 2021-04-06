// Copyright (c) 2021 akachain
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package mock

import (
	"encoding/json"
	"github.com/hyperledger/fabric/common/metrics/disabled"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/statecouchdb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
	"github.com/hyperledger/fabric/core/ledger/util/couchdb"
	"github.com/spf13/viper"

	"io/ioutil"
	"time"
)

const (
	// The couchDB test will have this name: DefaultChannelName_chaincodeName
	DefaultChannelName   = "testchannel"   // Fabric channel
)

// CouchDBHandler holds 1 parameter:
// dbEngine: a VersionedDB object that is used by the chaincode to query.
// This is to guarantee that the test uses the same logic in interaction with stateDB as the chaincode.
// This also includes how chaincode builds its query to interact with the stateDB.
type CouchDBHandler struct {
	dbEngine *statecouchdb.VersionedDB
	chaincodeName string
}

func getCouchDBConfig() *couchdb.Config {
	// statedb.VersionedDB does not publish its couchDB object
	// Thus, we'll have to recreate the set required config data to use state couchdb
	redoPath, _ := ioutil.TempDir("", "redoPath")
	conf := &couchdb.Config{
		Address:             viper.GetString("ledger.state.couchDBConfig.couchDBAddress"),
		Username:            viper.GetString("ledger.state.couchDBConfig.username"),
		Password:            viper.GetString("ledger.state.couchDBConfig.password"),
		InternalQueryLimit:  1000,
		MaxBatchUpdateSize:  1000,
		MaxRetries:          3,
		MaxRetriesOnStartup: 20,
		RequestTimeout:      35 * time.Second,
		RedoLogPath:         redoPath,
		UserCacheSizeMBs:    8,
	}

	return conf
}

// NewCouchDBHandler returns a new CouchDBHandler and setup database for testing
func NewCouchDBHandler(isDrop bool, ccName string) (*CouchDBHandler, error) {

	// Sometimes we'll have to drop the database to clean all previous test
	if isDrop == true {
		cleanUp(ccName)
	}

	// Create a new dbEngine for the channel
	config := getCouchDBConfig()
	couchState, _ := statecouchdb.NewVersionedDBProvider(config, &disabled.Provider{}, &statedb.Cache{})

	// This step creates a redundant meta database with name channel_ ,
	// there should be some ways to prevent this. We leave it for now
	h, err := couchState.GetDBHandle(DefaultChannelName)
	if err != nil {
		return nil, err
	}

	// now init the dbHandler with our couchdb engine
	handler := new(CouchDBHandler)
	handler.dbEngine = h.(*statecouchdb.VersionedDB)
	handler.chaincodeName = ccName
	return handler, nil
}

func cleanUp(ccName string) error {
	config := getCouchDBConfig()
	ins, er := couchdb.CreateCouchInstance(config, &disabled.Provider{})
	if er != nil {
		return er
	}
	dbName := couchdb.ConstructNamespaceDBName(DefaultChannelName, ccName)
	db := couchdb.CouchDatabase{CouchInstance: ins, DBName: dbName}
	_, er = db.DropDatabase()
	return er
}

// SaveDocument stores a value in couchDB
func (handler *CouchDBHandler) SaveDocument(key string, value []byte) error {
	// unmarshal the value param
	var doc map[string]interface{}
	json.Unmarshal(value, &doc)

	// Save the doc in database
	batch := statedb.NewUpdateBatch()
	batch.Put(handler.chaincodeName, key, value, version.NewHeight(1, 1))
	savePoint := version.NewHeight(1, 2)
	err := handler.dbEngine.ApplyUpdates(batch, savePoint)

	return err
}

// QueryDocument executes a query string and return results
func (handler *CouchDBHandler) QueryDocument(query string) (statedb.ResultsIterator, error) {
	rs, er := handler.dbEngine.ExecuteQuery(handler.chaincodeName, query)
	return rs, er
}

// QueryDocumentWithPagination executes a query string and return results
func (handler *CouchDBHandler) QueryDocumentWithPagination(query string, limit int32, bookmark string) (statedb.ResultsIterator, error) {
	queryOptions := make(map[string]interface{})
	if limit != 0 {
		queryOptions["limit"] = limit
	}
	if bookmark != "" {
		queryOptions["bookmark"] = bookmark
	}
	rs, er := handler.dbEngine.ExecuteQueryWithMetadata(handler.chaincodeName, query, queryOptions)
	return rs, er
}

// ReadDocument executes a query string and return results
func (handler *CouchDBHandler) ReadDocument(id string) ([]byte, error) {
	rs, er := handler.dbEngine.GetState(handler.chaincodeName, id)
	if er != nil {
		return nil, er
	}
	// found no document in db with id
	if rs == nil {
		return nil, nil
	}
	return rs.Value, er
}

// QueryDocumentByRange get a list of documents from couchDB by key range
func (handler *CouchDBHandler) QueryDocumentByRange(startKey, endKey string) (statedb.ResultsIterator, error) {
	rs, er := handler.dbEngine.GetStateRangeScanIterator(handler.chaincodeName, startKey, endKey)
	return rs, er
}

//// QueryDocumentByRange get a list of documents from couchDB by key range
//// TODO: GetStateRangeScanIteratorWithMetadata does not accept bookmark
//func (handler *CouchDBHandler) QueryDocumentByRangeWithPagination(startKey, endKey string, limit int32, bookmark string) (statedb.ResultsIterator, error) {
//	queryOptions := make(map[string]interface{})
//	if limit != 0 {
//		queryOptions["limit"] = limit
//	}
//	//if bookmark != "" {
//	//	queryOptions["bookmark"] = bookmark
//	//}
//
//	rs, er := handler.dbEngine.GetStateRangeScanIteratorWithMetadata(handler.chaincodeName, startKey, endKey, queryOptions)
//	return rs, er
//}

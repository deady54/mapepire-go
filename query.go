package mapepire

import (
	"fmt"
	"sync"
)

// Represents options for query execution
type QueryOptions struct {
	Rows        int      // The amount of rows to fetch
	Parameters  []string // Parameters, if any
	TerseResult bool     // Whether the result returns in terse format
	CLcommand   bool     // Whether the command is a CL command
}

// Represents a SQL Query that can be executed and managed within a SQL job
type Query struct {
	id          string  // The unique identifier
	clCommand   string  // CL command
	sqlQuery    string  // SQL query
	parameters  string  // Parameters, if any
	terse       bool    // Whether the result returns in terse format
	rowsToFetch string  // The amount of rows to fetch
	prepared    bool    // Whether the query has been prepared
	job         *SQLJob // Pointer to the SQL Job
	state       int     // The current state of the query
}

// Represents a query list managed by the job
type queryList struct {
	list []*Query   // List of all open queries
	lock sync.Mutex // Mutex
}

const (
	STATE_RUN_DONE = iota
	STATE_RUN_MORE_DATA
	STATE_NOT_YET_RUN
)

// Receive a new query list
func newQueryList() *queryList {
	return &queryList{list: []*Query{}}
}

// Add a new Query to the list
func (ql *queryList) addQuery(query *Query) {
	ql.lock.Lock()
	ql.list = append(ql.list, query)
	ql.lock.Unlock()
}

// validates the cont_id
func (ql *queryList) validateID(ID string) bool {
	ql.lock.Lock()
	defer ql.lock.Unlock()

	for _, query := range ql.list {
		if query.id == ID && query.state != STATE_RUN_DONE {
			return true
		}
	}
	return false
}

// cleans up the query list
func (ql *queryList) cleanup() {
	ql.lock.Lock()
	defer ql.lock.Unlock()

	newList := make([]*Query, 0, len(ql.list))
	for _, query := range ql.list {
		if query.state != STATE_RUN_DONE {
			newList = append(newList, query)
		}
	}
	ql.list = newList
}

// Run CL command
func (q *Query) RunCL() *ServerResponse {
	if q.clCommand == "" {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("need CL command")}
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	q.prepared = false

	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"cl","cmd":"%s"}`, q.id, q.clCommand)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	return q.sendRequest(request)
}

// Run SQL query
func (q *Query) RunSQL() *ServerResponse {
	if q.sqlQuery == "" {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("need SQL")}
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	q.prepared = false

	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"sql","rows":"%s","sql":"%s","terse":%t}`, q.id, q.rowsToFetch, q.sqlQuery, q.terse)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	return q.sendRequest(request)
}

// Prepare a SQL query
func (q *Query) PrepareSQL() error {
	if q.sqlQuery == "" {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return fmt.Errorf("need SQL")
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	q.prepared = true

	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"prepare_sql","sql":"%s","terse":%t}`, q.id, q.sqlQuery, q.terse)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	resp := q.job.send(*request)
	if resp.Error != nil {
		return resp.Error
	}

	q.job.setJobStatus(JOBSTATUS_READY)
	return nil
}

// Execute a prepared SQL query with the ID
func (q *Query) Execute(contID string) *ServerResponse {

	valid := q.job.queryList.validateID(contID)
	if !valid {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("need ID from previous SQL")}
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)

	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"execute","cont_id":"%s","parameters":%s,"rows":"%s","terse":%t}`, q.id, contID, q.parameters, q.rowsToFetch, q.terse)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	return q.sendRequest(request)
}

// Prepare and execute the SQL statement
func (q *Query) PrepareSQL_Execute() *ServerResponse {

	if q.sqlQuery == "" {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("need SQL")}
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	q.prepared = true

	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"prepare_sql_execute","sql":"%s","parameters":%s,"rows":"%s","terse":%t}`, q.id, q.sqlQuery, q.parameters, q.rowsToFetch, q.terse)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	return q.sendRequest(request)
}

// Fetch more rows from a previous request with the ID
func (q *Query) SQLmore(contID string, rows string) *ServerResponse {

	valid := q.job.queryList.validateID(contID)
	if !valid {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("need ID from previous SQL")}
	} else if q.state == STATE_NOT_YET_RUN {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return &ServerResponse{Error: fmt.Errorf("statement has not yet been run")}
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	jsonreq :=
		fmt.Sprintf(`{"id":"%s","type":"sqlmore","cont_id":"%s","rows":"%s"}`, q.id, contID, rows)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	response := q.sendRequest(request)
	if response.Success {
		response.HasResults = true
	}
	return response
}

// Close cursor from a previous request
func (q *Query) SQLClose(contID string) error {

	valid := q.job.queryList.validateID(contID)
	if !valid {
		q.job.setJobStatus(JOBSTATUS_ERROR)
		return fmt.Errorf("need ID from previous SQL")
	}

	q.job.setJobStatus(JOBSTATUS_BUSY)
	jsonreq :=
		fmt.Sprintf(`{"id":"%v","type":"sqlclose","cont_id":"%v"}`, q.id, contID)

	request := &serverRequest{
		id:      q.id,
		jsonreq: jsonreq,
	}

	resp := q.job.send(*request)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// sends the request and sets the query state
func (q *Query) sendRequest(request *serverRequest) *ServerResponse {
	resp := q.job.send(*request)
	if resp.Error != nil {
		return resp
	}

	if resp.IsDone && resp.Success {
		q.state = STATE_RUN_DONE
	} else if resp.Success && !resp.IsDone {
		q.state = STATE_RUN_MORE_DATA
	}
	q.job.queryList.cleanup()

	q.job.setJobStatus(JOBSTATUS_READY)
	return resp
}

package mapepire

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Represents a DB2 server daemon with connection details.
type DaemonServer struct {
	Host               string // Hostname or IP
	User               string // Username for authentication
	Password           string // Password for authentication
	Port               string // Default port is 8076
	IgnoreUnauthorized bool   // Ignore unauthorized certificate
	Technique          string // CLI or TCP
	Properties         string // A semicolon-delimited list of JDBC connection properties
}

// Represents a SQL job that manages connections and queries to a database.
type SQLJob struct {
	ID         string          // Unique identifier
	Jobname    string          // Name of the Job
	Status     string          // Status of the Job
	Options    *TraceOptions   // Trace configuration options
	query      *Query          // Pointer to the query
	queryList  *queryList      // List of all open queries
	connection *websocket.Conn // Websocket connection
	counter    atomic.Uint32   // Atomic counter
	writeMutex sync.Mutex      // Mutex
}

const (
	JOBSTATUS_BUSY        = "BUSY"
	JOBSTATUS_ERROR       = "ERROR"
	JOBSTATUS_ENDED       = "ENDED"
	JOBSTATUS_READY       = "READY"
	JOBSTATUS_CONNECTING  = "CONNECTING"
	JOBSTATUS_NOT_STARTED = "NOT_STARTED"
)
const (
	writeErr = "error writing message: %v"
	readErr  = "error reading message: %v"
	jsonErr  = "error unmarshalling json: %v"
)

// Receive a new SQL job with the given ID
func NewSQLJob(ID string) *SQLJob {
	list := newQueryList()
	return &SQLJob{ID: ID, Status: JOBSTATUS_NOT_STARTED, queryList: list}
}

// Creates a websocket connection and connects to the server.
func (s *SQLJob) Connect(server DaemonServer) error {

	s.setJobStatus(JOBSTATUS_CONNECTING)
	if server.Port == "" {
		server.Port = "8076"
	}

	url := fmt.Sprintf("wss://%s:%s/db/", server.Host, server.Port)
	dialer := *websocket.DefaultDialer

	if server.IgnoreUnauthorized {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	header := http.Header{}
	header.Add("authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(server.User+":"+server.Password)))

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return fmt.Errorf("error connecting to websocket: %v", err)
	}
	s.connection = conn

	var jsonreq string
	if server.Technique != "" {
		jsonreq =
			fmt.Sprintf(`{"id":"%v","type":"connect","techinque":"%v","props":"%v"}`, s.ID, server.Technique, server.Properties)
	} else {
		jsonreq =
			fmt.Sprintf(`{"id":"%v","type":"connect","props":"%v"}`, s.ID, server.Properties)
	}

	request := serverRequest{
		id:      s.ID,
		jsonreq: jsonreq,
	}

	response := s.send(request)
	if response.Error != nil {
		return response.Error
	}

	s.Jobname, response.Error = s.getDBJob()
	if response.Error != nil {
		return response.Error
	}
	s.setJobStatus(JOBSTATUS_READY)
	return nil
}

// sends requests to the server
func (s *SQLJob) send(req serverRequest) *ServerResponse {

	response := &ServerResponse{
		ID: req.id,
	}

	if s.connection == nil {
		return s.setError(response, "Error: %v", fmt.Errorf("need a websocket connection"))
	}

	s.writeMutex.Lock()
	if err := s.connection.WriteMessage(1, []byte(req.jsonreq)); err != nil {
		return s.setError(response, writeErr, err)
	}

	_, resp, err := s.connection.ReadMessage()
	if err != nil {
		return s.setError(response, readErr, err)
	}
	s.writeMutex.Unlock()

	response.SqlRC, response.SqlState, response.Error = checkJsonErr(resp)
	if response.Error != nil {
		return response
	}

	// Only works if data is received in terse format
	if s.query != nil && s.query.terse {
		resp = []byte(strings.Replace(string(resp), `"data":[[`, `"terse_data":[[`, 1))
	}

	if err := json.Unmarshal(resp, response); err != nil {
		return s.setError(response, jsonErr, err)
	}

	if s.Jobname != "" {
		response.Job = s.Jobname
	}

	return response
}

// checks JSON for errors
func checkJsonErr(jsonres []byte) (int, string, error) {
	var checkError struct {
		Error    string
		SqlRC    int    `json:"sql_rc"`
		SqlState string `json:"sql_state"`
	}

	json.Unmarshal(jsonres, &checkError)
	if checkError.Error != "" || checkError.SqlState != "" {
		return checkError.SqlRC, checkError.SqlState, fmt.Errorf(checkError.Error)
	}

	return 0, "", nil
}

// Creates a query with the SQL
func (s *SQLJob) Query(sql string) (*Query, error) {
	return s.QueryWithOptions(sql, QueryOptions{})
}

// Creates a query with the CL command
func (s *SQLJob) ClCommand(command string) (*Query, error) {
	return s.QueryWithOptions(command, QueryOptions{IsCLcommand: true})
}

// Creates a query with the given options
func (s *SQLJob) QueryWithOptions(command string, options QueryOptions) (*Query, error) {

	if command == "" {
		return nil, fmt.Errorf("SQL or CL command required")
	}

	jsonParams, err := func() (string, error) {
		if len(options.Parameters) == 1 {
			param, err := json.Marshal(options.Parameters[0])
			return string(param), err
		} else {
			params, err := json.Marshal(options.Parameters)
			return string(params), err
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("error marshalling to JSON: %v", err)
	}

	ID := s.getNewUniqueID()

	var rows string
	if options.Rows > 0 {
		rows = fmt.Sprint(options.Rows)
	}

	query := &Query{
		ID:          ID,
		parameters:  jsonParams,
		rowsToFetch: rows,
		terse:       options.TerseResult,
		job:         s,
	}
	query.state.Store(STATE_NOT_YET_RUN)

	if options.Parameters != nil {
		query.prepared = true
	}

	if options.IsCLcommand {
		query.clCommand = command
	} else {
		query.sqlQuery = command
	}

	s.queryList.addQuery(query)

	s.query = query
	return query, nil
}

// Set trace configuration options
func (s *SQLJob) SetTraceConfig(ops TraceOptions) error {
	var jsonreq string

	var allFields = ops.Tracelevel != "" && ops.Tracedest != "" && ops.Jtopentracedest != "" && ops.Jtopentracelevel != ""
	var jtFields = ops.Jtopentracelevel != "" && ops.Jtopentracedest != ""
	var traceFields = ops.Tracelevel != "" && ops.Tracedest != ""

	if allFields {
		jsonreq =
			fmt.Sprintf(`{"id":"%s","type":"setconfig","jtopentracelevel":"%s","jtopentracedest":"%s","tracelevel":"%s","tracedest":"%s"}`, s.ID, ops.Jtopentracelevel, ops.Jtopentracedest, ops.Tracelevel, ops.Tracedest)

	} else if traceFields {
		jsonreq =
			fmt.Sprintf(`{"id":"%s","type":"setconfig","tracelevel":"%s","tracedest":"%s"}`, s.ID, ops.Tracelevel, ops.Tracedest)

	} else if jtFields {
		jsonreq =
			fmt.Sprintf(`{"id":"%s","type":"setconfig","tracelevel":"%s","tracedest":"%s"}`, s.ID, ops.Jtopentracelevel, ops.Jtopentracedest)
	} else {
		return fmt.Errorf("need atleast 2 fields; level and dest of the same tracer")
	}

	s.writeMutex.Lock()
	err := s.connection.WriteMessage(1, []byte(jsonreq))
	if err != nil {
		return fmt.Errorf(writeErr, err)
	}

	_, resp, err := s.connection.ReadMessage()
	if err != nil {
		return fmt.Errorf(readErr, err)
	}
	_, _, err = checkJsonErr(resp)
	if err != nil {
		return err
	}
	s.writeMutex.Unlock()

	trace := &TraceOptions{}
	err = json.Unmarshal(resp, trace)
	if err != nil {
		return fmt.Errorf(jsonErr, err)
	}

	ops.tracing = true
	s.Options = &ops
	return nil
}

// Receive trace data (after setting config)
func (s *SQLJob) GetTraceData() error {
	if s.Options == nil || !s.Options.tracing {
		return fmt.Errorf("need to set the trace config")
	}

	jsonreq := fmt.Sprintf(`{"id":"%v","type":"gettracedata"}`, s.ID)

	s.writeMutex.Lock()
	err := s.connection.WriteMessage(1, []byte(jsonreq))
	if err != nil {
		return fmt.Errorf(writeErr, err)
	}

	_, resp, err := s.connection.ReadMessage()
	if err != nil {
		return fmt.Errorf(readErr, err)
	}
	s.writeMutex.Unlock()

	_, _, err = checkJsonErr(resp)
	if err != nil {
		return err
	}

	var data traceData
	data.Tracedest = s.Options.Tracedest
	data.Jtopentracedest = s.Options.Jtopentracedest

	err = json.Unmarshal(resp, &data)
	if err != nil {
		return fmt.Errorf(jsonErr, err)
	}

	err = createTraceFile(data)
	if err != nil {
		return err
	}

	return nil
}

// Creates a file for the trace data
func createTraceFile(t traceData) error {
	if t.Tracedest == "file" || t.Tracedest == "FILE" {
		filename := "trace-" + time.Now().Format("2002-01-02") + ".html"
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}

		_, err = file.Write([]byte(t.Tracedata))
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}
	if t.Jtopentracedest == "file" || t.Jtopentracedest == "FILE" {
		filename := "jtopentrace-" + time.Now().Format("2002-01-02") + ".txt"
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}

		_, err = file.Write([]byte(t.Jtopentracedata))
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}

// Closes the SQL job and websocket connection.
func (s *SQLJob) Close() error {
	if s.connection == nil {
		return fmt.Errorf("no connection found")
	}

	s.setJobStatus(JOBSTATUS_ENDED)
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	err := s.connection.WriteMessage(1, []byte(`{"id":"bye","type":"exit"}`))
	if err != nil {
		return fmt.Errorf(writeErr, err)
	}
	err = s.connection.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %v", err)
	}

	s.connection = nil
	s.Options = nil
	s.query = nil
	return nil
}

// Receive the current job status
func (s *SQLJob) GetStatus() string {
	return s.Status
}

// Receive the current version info
func (s *SQLJob) GetVersion() (string, error) {
	if s.connection == nil {
		return "", fmt.Errorf("no connection found")
	}
	jsonreq := `{"id":"versionCheck","type":"getversion"}`

	s.writeMutex.Lock()
	err := s.connection.WriteMessage(1, []byte(jsonreq))
	if err != nil {
		return "", fmt.Errorf(writeErr, err)
	}

	_, resp, err := s.connection.ReadMessage()
	if err != nil {
		return "", fmt.Errorf(readErr, err)
	}
	s.writeMutex.Unlock()

	var response struct{ Version string }
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return "", fmt.Errorf(jsonErr, err)
	}

	return response.Version, nil
}

// Receive the name of the Job
func (s *SQLJob) getDBJob() (string, error) {
	if s.ID == "" {
		return "", fmt.Errorf("need a job ID")
	}

	jsonreq := fmt.Sprintf(`{"id":"%v","type":"getdbjob"}`, s.ID)

	request := &serverRequest{
		id:      s.ID,
		jsonreq: jsonreq,
	}

	resp := s.send(*request)
	if resp.Error != nil {
		return "", resp.Error
	}
	s.Jobname = resp.Job

	resp.IsDone = true
	return resp.Job, nil
}

func (s *SQLJob) getNewUniqueID() string {
	count := s.counter.Add(1)
	ID := strconv.Itoa(int(count))
	return ID
}

// Set errors
func (s *SQLJob) setError(resp *ServerResponse, info string, err error) *ServerResponse {
	s.setJobStatus(JOBSTATUS_ERROR)
	resp.Error = fmt.Errorf(info, err)
	return resp
}

// Set the current job status
func (s *SQLJob) setJobStatus(status string) {
	s.Status = status
}

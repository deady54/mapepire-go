package mapepire

import (
	"fmt"
	"log"
	"testing"
)

func initSQLTable(command string, queryops2 QueryOptions) (*SQLJob, *Query) {
	job := NewSQLJob("test")
	err := job.Connect(server)
	if err != nil {
		log.Fatal(err)
	}

	query, _ := job.Query("CREATE TABLE qtemp.TEMPTEST (ID CHAR(8) NOT NULL, DESCRIPTION VARCHAR(60) NOT NULL, SERIALNO CHAR(12) NOT NULL)")
	resp := query.RunSQL()
	if resp.Error != nil {
		log.Fatal(resp.Error)
	}

	insertquery, _ := job.Query(`INSERT INTO TEMPTEST VALUES (1, 'Lorem ipsum', 121212),
	(2, 'dolor sit amet', 232323),
	(3, 'consetetur sadipscing elitr', 343434),
	(4, 'sed diam nonumy', 454545),
	(5, 'eirmod tempor', 565656)`)
	resp2 := insertquery.RunSQL()
	log.Println(resp2)

	query2, err := job.QueryWithOptions(command, queryops2)
	if err != nil {
		log.Fatal(err)
	}
	return job, query2
}

// Run CL valid
func TestRunCL(t *testing.T) {
	queryops := QueryOptions{
		CLcommand: true,
	}

	job := NewSQLJob("test")
	job.Connect(server)

	command := fmt.Sprintf("DSPUSRPRF %s", server.User)
	query, err := job.QueryWithOptions(command, queryops)
	if err != nil {
		t.Errorf("should not throw error")
	}

	have := query.RunCL()

	want := &ServerResponse{
		ID:      "1",
		Success: true,
		Error:   nil,
		IsDone:  true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
}

// Invalid CL command
func TestRunCLinvalid(t *testing.T) {
	queryops := QueryOptions{
		CLcommand: true,
	}
	job := NewSQLJob("test")
	job.Connect(server)

	query, err := job.QueryWithOptions("INVALIDCOMMAND", queryops)
	if err != nil {
		t.Errorf("should not throw error")
	}

	have := query.RunCL()

	want := &ServerResponse{
		ID:       "1",
		Success:  false,
		SqlState: "42806",
		SqlRC:    -303,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.SqlState != want.SqlState {
		t.Errorf("got %v, want %v", have.SqlState, want.SqlState)
	}
	if have.SqlRC != want.SqlRC {
		t.Errorf("got %v, want %v", have.SqlRC, want.SqlRC)
	}
}

// CL invalid 2
func TestRunClinvalid2(t *testing.T) {
	queryops := QueryOptions{}
	job := NewSQLJob("test")
	job.Connect(server)

	query, err := job.QueryWithOptions("SELECT * FROM TEMPTEST", queryops)
	if err != nil {
		t.Errorf("should not throw error")
	}
	have := query.RunCL()

	if have.Error == nil {
		t.Errorf("should throw error")
	}
}

// Run SQL valid
func TestRunSQL(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	have := query.RunSQL()

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Run SQL in terse format
func TestRunSQLTerse(t *testing.T) {
	queryops := QueryOptions{
		Rows:        5,
		TerseResult: true,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	have := query.RunSQL()

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should be in Terse_data")
	}
	if have.TerseData == nil {
		t.Errorf("should have Terse_data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Run SQL invalid
func TestRunSQLInvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM INVALIDTABLE", queryops)
	have := query.RunSQL()

	want := &ServerResponse{
		ID:         "3",
		Success:    false,
		IsDone:     false,
		HasResults: false,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw Error")
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should not have data")
	}
	if have.Metadata != nil {
		t.Errorf("should not have metadata")
	}
}

// Run SQL Insert
func TestRunSQLInsert(t *testing.T) {
	job := NewSQLJob("test")
	job.Connect(server)

	query, _ := job.Query("DECLARE GLOBAL TEMPORARY TABLE TEMPTEST (ID CHAR(8) NOT NULL, DESCRIPTION VARCHAR(60) NOT NULL, SERIALNO CHAR(12) NOT NULL)")
	resp := query.RunSQL()
	if resp.Error != nil {
		log.Fatal(resp.Error)
	}

	insertquery, _ := job.Query("INSERT INTO TEMPTEST VALUES (1, 'testtest', 121212)")
	have := insertquery.RunSQL()

	want := ServerResponse{
		ID:          "2",
		Success:     true,
		HasResults:  false,
		UpdateCount: 1,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != nil {
		t.Errorf("should not throw Error")
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("have %v, want %v", have.HasResults, want.HasResults)
	}
}

// Run SQL Update
func TestRunSQLUpdate(t *testing.T) {
	queryops := QueryOptions{}

	job, query := initSQLTable("UPDATE TEMPTEST SET ID = 5, DESCRIPTION = 'test', SERIALNO = 545454 WHERE ID = 1", queryops)
	update := query.RunSQL()
	if update.Error != nil {
		t.Errorf("should not throw error")
	}

	query2, _ := job.Query("SELECT * FROM TEMPTEST WHERE ID = 5")
	have := query2.RunSQL()
	want := ServerResponse{
		ID:          "4",
		Success:     true,
		HasResults:  true,
		IsDone:      true,
		UpdateCount: 1,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != nil {
		t.Errorf("should not throw Error")
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
	if update.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
}

// Run SQL Delete
func TestRunSQLDelete(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("DELETE FROM TEMPTEST WHERE ID = 1", queryops)
	have := query.RunSQL()

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		Error:       nil,
		UpdateCount: 1,
	}

	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Error != want.Error {
		t.Errorf("should not throw error")
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
}

// Prepare SQL valid
func TestPrepareSQL(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	have := query.PrepareSQL()

	if have != nil {
		t.Errorf("should not throw error")
	}
}

// Prepare SQL terse format
func TestPrepareSQLTerse(t *testing.T) {
	queryops := QueryOptions{
		Rows:        5,
		TerseResult: true,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	have := query.PrepareSQL()

	if have != nil {
		t.Errorf("should not throw error")
	}
}

// Prepare SQL invalid
func TestPrepareSQLInvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM INVALIDTABLE", queryops)
	have := query.PrepareSQL()

	if have == nil {
		t.Errorf("should throw error")
	}
}

// Prepare SQL Insert
func TestPrepareSQLInsert(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("INSERT INTO TEMPTEST VALUES (2, 'something', 545454)", queryops)
	have := query.PrepareSQL()

	if have != nil {
		t.Errorf("should not throw error")
	}
}

// Prepare SQL Update
func TestPrepareSQLUpdate(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("UPDATE TEMPTEST SET ID = 5, DESCRIPTION = 'test', SERIALNO = 545454 WHERE ID = 1", queryops)
	have := query.PrepareSQL()

	if have != nil {
		t.Errorf("should not throw error")
	}
}

// Prepare SQL Delete
func TestPrepareSQLDelete(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("DELETE FROM TEMPTEST WHERE ID = 1", queryops)
	have := query.PrepareSQL()

	if have != nil {
		t.Errorf("should not throw error")
	}
}

// Execute valid
func TestExecute(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	query.PrepareSQL()
	have := query.Execute(query.id)

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Execute SQL in terse format
func TestExecuteTerse(t *testing.T) {
	queryops := QueryOptions{
		Rows:        5,
		TerseResult: true,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	query.PrepareSQL()
	have := query.Execute(query.id)

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}

	if have.Data != nil {
		t.Errorf("should be in Terse_data")
	}
	if have.TerseData == nil {
		t.Errorf("should have Terse_data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Execute SQL invalid
func TestExecuteInvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM INVALIDTABLE", queryops)

	query.PrepareSQL()
	have := query.Execute(query.id)

	want := &ServerResponse{
		ID:         "3",
		Success:    false,
		IsDone:     false,
		HasResults: false,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should not have data")
	}
	if have.Metadata != nil {
		t.Errorf("should not have metadata")
	}
}

func TestExecuteInvalid2(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	query.PrepareSQL()
	have := query.Execute("invalid")

	want := &ServerResponse{
		Success:    false,
		IsDone:     false,
		HasResults: false,
	}

	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should not have data")
	}
	if have.Metadata != nil {
		t.Errorf("should not have metadata")
	}
}

// Execute SQL Insert
func TestExecuteInsert(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("INSERT INTO TEMPTEST VALUES (2, 'something', 545454)", queryops)
	query.PrepareSQL()
	have := query.Execute(query.id)

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		UpdateCount: 1,
		HasResults:  false,
		IsDone:      true,
	}

	if have.Error != nil {
		t.Errorf("should not throw error")
	}
	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("have %v, want %v", have.HasResults, want.HasResults)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("have %v, want %v", have.IsDone, want.IsDone)
	}
}

// Execute SQL Update
func TestExecuteUpdate(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("UPDATE TEMPTEST SET ID = 5, DESCRIPTION = 'test', SERIALNO = 545454 WHERE ID = 1", queryops)
	query.PrepareSQL()
	have := query.Execute(query.id)

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		UpdateCount: 1,
		HasResults:  false,
		IsDone:      true,
	}

	if have.Error != nil {
		t.Errorf("should not throw error")
	}
	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("have %v, want %v", have.HasResults, want.HasResults)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("have %v, want %v", have.IsDone, want.IsDone)
	}
}

// Execute SQL Delete
func TestExecuteDelete(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("DELETE FROM TEMPTEST WHERE ID = 1", queryops)
	query.PrepareSQL()
	have := query.Execute(query.id)

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		Error:       nil,
		UpdateCount: 1,
	}

	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Error != want.Error {
		t.Errorf("should not throw error")
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
}

// Prepare and execute valid
func TestPrepareSQL_Execute(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	have := query.PrepareSQL_Execute()

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Prepare and Execute SQL in terse format
func TestPrepareSQL_ExecuteTerse(t *testing.T) {
	queryops := QueryOptions{
		Rows:        5,
		TerseResult: true,
	}

	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	have := query.PrepareSQL_Execute()

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should be in Terse_data")
	}
	if have.TerseData == nil {
		t.Errorf("should have Terse_data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// Prepare and Execute invalid SQL
func TestPrepareSQL_ExecuteInvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}

	_, query := initSQLTable("SELECT * FROM INVALIDTABLE", queryops)

	have := query.PrepareSQL_Execute()

	want := &ServerResponse{
		ID:         "3",
		Success:    false,
		IsDone:     false,
		HasResults: false,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should not have data")
	}
	if have.Metadata != nil {
		t.Errorf("should not have metadata")
	}
}

// Prepare and Execute SQL Insert
func TestPrepareSQL_ExecuteInsert(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("INSERT INTO TEMPTEST VALUES (2, 'something', 545454)", queryops)
	have := query.PrepareSQL_Execute()

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		UpdateCount: 1,
		HasResults:  false,
		IsDone:      true,
	}

	if have.Error != nil {
		t.Errorf("should not throw error")
	}
	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("have %v, want %v", have.HasResults, want.HasResults)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("have %v, want %v", have.IsDone, want.IsDone)
	}
}

// Prepare and Execute SQL Update
func TestPrepareSQL_ExecuteUpdate(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("UPDATE TEMPTEST SET ID = 5, DESCRIPTION = 'test', SERIALNO = 545454 WHERE ID = 1", queryops)
	have := query.PrepareSQL_Execute()

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		UpdateCount: 1,
		HasResults:  false,
		IsDone:      true,
	}

	if have.Error != nil {
		t.Errorf("should not throw error")
	}
	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("have %v, want %v", have.HasResults, want.HasResults)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("have %v, want %v", have.IsDone, want.IsDone)
	}
}

// Prepare and Execute SQL Delete
func TestPrepareSQL_ExecuteDelete(t *testing.T) {
	queryops := QueryOptions{}

	_, query := initSQLTable("DELETE FROM TEMPTEST WHERE ID = 1", queryops)
	have := query.PrepareSQL_Execute()

	want := ServerResponse{
		ID:          "3",
		Success:     true,
		Error:       nil,
		UpdateCount: 1,
	}

	if have.ID != want.ID {
		t.Errorf("have %v, want %v", have.ID, want.ID)
	}
	if have.Error != want.Error {
		t.Errorf("should not throw error")
	}
	if have.Success != want.Success {
		t.Errorf("have %v, want %v", have.Success, want.Success)
	}
	if have.UpdateCount != want.UpdateCount {
		t.Errorf("have %v, want %v", have.UpdateCount, want.UpdateCount)
	}
}

// Prepare and Execute SQL with Params
func TestPrepareSQL_ExecuteParam(t *testing.T) {
	queryops := QueryOptions{
		Parameters: []string{"3"},
		Rows:       5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST WHERE ID = ?", queryops)

	have := query.PrepareSQL_Execute()

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     true,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
	if have.Metadata == nil {
		t.Errorf("should have metadata")
	}
}

// SQLmore valid
func TestSQLmore(t *testing.T) {
	queryops := QueryOptions{
		Rows: 1,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	query.PrepareSQL_Execute()
	have := query.SQLmore(query.id, "5")

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     true,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %v, want %v", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data == nil {
		t.Errorf("should have data")
	}
}

// SQL more in terse format
func TestSQLmoreTerse(t *testing.T) {
	queryops := QueryOptions{
		Rows:        1,
		TerseResult: true,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)

	query.PrepareSQL_Execute()
	have := query.SQLmore(query.id, "1")

	want := &ServerResponse{
		ID:         "3",
		Success:    true,
		IsDone:     false,
		Error:      nil,
		HasResults: true,
	}

	if have.ID != want.ID {
		t.Errorf("got %s, want %s", have.ID, want.ID)
	}
	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error != want.Error {
		t.Errorf("got %s, want %s", have.Error, want.Error)
	}
	if have.IsDone != want.IsDone {
		t.Errorf("got %t, want %t", have.IsDone, want.IsDone)
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
	if have.Data != nil {
		t.Errorf("should be in Terse_data")
	}
	if have.TerseData == nil {
		t.Errorf("should have Terse_data")
	}
}

func TestSQLmoreinvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 1,
	}

	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	query.PrepareSQL_Execute()
	have := query.SQLmore("invalid", "1")

	want := &ServerResponse{
		Success:    false,
		HasResults: false,
	}

	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
}

// SQL more invalid 2
func TestSQLmoreInvalid2(t *testing.T) {
	queryops := QueryOptions{
		Rows: 1,
	}

	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	query.PrepareSQL_Execute()
	have := query.SQLmore("", "1")

	want := &ServerResponse{
		Success:    false,
		HasResults: false,
	}

	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
}

func TestSQLmoreinvalid3(t *testing.T) {
	queryops := QueryOptions{
		Rows: 1,
	}

	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	have := query.SQLmore("5", "1")

	want := &ServerResponse{
		Success:    false,
		HasResults: false,
	}

	if have.Success != want.Success {
		t.Errorf("got %t, want %t", have.Success, want.Success)
	}
	if have.Error == nil {
		t.Errorf("should throw error")
	}
	if have.HasResults != want.HasResults {
		t.Errorf("got %v, want %v", have.HasResults, want.HasResults)
	}
}

// SQL Close
func TestSQLClose(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	query.RunSQL()
	err := query.SQLClose(query.id)
	if err != nil {
		t.Errorf("should not throw error")
	}
}

// SQLClose null
func TestSQLCloseNULL(t *testing.T) {
	queryops := QueryOptions{
		Rows: 5,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	query.RunSQL()
	err := query.SQLClose("")
	if err == nil {
		t.Errorf("should throw error")
	}
}

// SQLClose invalid
func TestSQLCloseInvalid(t *testing.T) {
	queryops := QueryOptions{
		Rows: 50,
	}
	_, query := initSQLTable("SELECT * FROM TEMPTEST", queryops)
	query.RunSQL()
	err := query.SQLClose("invalid")
	if err == nil {
		t.Errorf("should throw error")
	}
}

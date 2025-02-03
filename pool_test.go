package mapepire

import (
	"log"
	"testing"
)

func initPoolSQLTable(pool *JobPool) error {
	_, err := pool.ExecuteSQL("CREATE TABLE qtemp.TEMPTEST (ID decimal(8) NOT NULL, DESCRIPTION VARCHAR(60) NOT NULL, SERIALNO CHAR(12) NOT NULL)")
	if err != nil {
		return err
	}
	_, err = pool.ExecuteSQL(`INSERT INTO TEMPTEST VALUES (1, 'Lorem ipsum', 121212),
	(2, 'dolor sit amet', 232323),
	(3, 'consetetur sadipscing elitr', 343434),
	(4, 'sed diam nonumy', 454545),
	(5, 'eirmod tempor', 565656)`)
	if err != nil {
		return err
	}
	return nil
}

func TestNewPool(t *testing.T) {
	_, err := NewPool(PoolOptions{Creds: server, MaxSize: 5, StartingSize: 2})
	if err != nil {
		t.Errorf("should not throw error")
	}
	_, err = NewPool(PoolOptions{Creds: server, MaxSize: 0, StartingSize: 3})
	if err == nil {
		t.Errorf("should throw error")
	}
	_, err = NewPool(PoolOptions{Creds: server, MaxSize: 5, StartingSize: 0})
	if err == nil {
		t.Errorf("should throw error")
	}
	_, err = NewPool(PoolOptions{Creds: server, MaxSize: 3, StartingSize: 5})
	if err == nil {
		t.Errorf("should throw error")
	}
}

func TestNewPoolJob(t *testing.T) {
	pool, err := NewPool(PoolOptions{Creds: server, StartingSize: 1, MaxSize: 2, MaxWaitTime: 1})
	if err != nil {
		t.Errorf("should not throw error")
	}
	pool.GetJob()
	have, err := pool.GetJob()
	if err != nil {
		t.Errorf("should not throw error")
	}

	if have.ID != "PoolJob 2" {
		t.Errorf("have %v, want 'PoolJob 2'", have.ID)
	}
}

func TestExecuteSQL(t *testing.T) {
	pool, err := NewPool(PoolOptions{Creds: server, MaxWaitTime: 3, MaxSize: 1, StartingSize: 1})
	if err != nil {
		t.Errorf("should not throw error")
	}
	err = initPoolSQLTable(pool)
	if err != nil {
		t.Errorf("should not throw error")
	}
	resp, err := pool.ExecuteSQLWithOptions("SELECT * FROM TEMPTEST", QueryOptions{Rows: 5})
	if err != nil {
		t.Errorf("should not throw error")
	}
	log.Println(resp)
	pool.Close()
}

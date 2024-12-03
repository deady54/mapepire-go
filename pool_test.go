package mapepire

import (
	"log"
	"testing"
)

func TestNewPool(t *testing.T) {
	_, err := NewPool(PoolOptions{Creds: &server, MaxSize: 5, StartingSize: 2})
	if err != nil {
		t.Errorf("should not throw error")
	}
	_, err = NewPool(PoolOptions{Creds: &server, MaxSize: 0, StartingSize: 3})
	if err == nil {
		t.Errorf("should throw error")
	}
	_, err = NewPool(PoolOptions{Creds: &server, MaxSize: 5, StartingSize: 0})
	if err == nil {
		t.Errorf("should throw error")
	}
	_, err = NewPool(PoolOptions{Creds: &server, MaxSize: 3, StartingSize: 5})
	if err == nil {
		t.Errorf("should throw error")
	}
}

func TestNewPoolJob(t *testing.T) {
	pool, err := NewPool(PoolOptions{Creds: &server, StartingSize: 1, MaxSize: 2, MaxWaitTime: 1})
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
	pool, err := NewPool(PoolOptions{Creds: &server, MaxWaitTime: 3, MaxSize: 5, StartingSize: 1})
	if err != nil {
		t.Errorf("should not throw error")
	}
	respChan, err := pool.ExecuteSQLWithOptions("SELECT * FROM DIPA5", QueryOptions{Rows: 5})
	if err != nil {
		t.Errorf("should not throw error")
	}
	log.Println(<-respChan)
	pool.Close()
}

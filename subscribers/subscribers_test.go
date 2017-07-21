package subscribers

import (
	"testing"
)

func TestNewSubscribers(t *testing.T) {
	s := NewSubscribers()
	if s == nil || s.subscribers == nil {
		t.Error("Error initializing Subscriber")
	}
}

func TestSubscriberGet(t *testing.T) {
	s := Subscribers{}
	s.subscribers = map[string]map[string]bool{"test": map[string]bool{"test1": true}}
	testmap := s.GetSubscriberList("test")
	if val, ok := testmap["test1"]; !ok {
		t.Error("Error retrieving map")
	} else if !val {
		t.Error("Error retrieveing Value")
	}
}

func TestSubscriberSet(t *testing.T) {
	s := Subscribers{subscribers: make(map[string]map[string]bool)}
	s.SetSubscriber("test", map[string]bool{"test1": true})
	if val, ok := s.subscribers["test"]; !ok {
		t.Error("Error setting map")
	} else if val2, ok := val["test1"]; !ok {
		t.Error("Error setting sub map")
	} else if !val2 {
		t.Error("Error Retrieving val")
	}
}

func TestSubscriberDel(t *testing.T) {
	s := Subscribers{}
	s.subscribers = map[string]map[string]bool{"test": map[string]bool{"test1": true}}
	if _, ok := s.subscribers["test"]; !ok {
		t.Error("Error in test Setup")
	} else {
		l1 := len(s.subscribers)
		s.DelSubscriberAll("test1")
		l2 := len(s.subscribers)
		if _, ok := s.subscribers["test"]; ok || l1 <= l2 {
			t.Error("Del method not working", l1, l2)
		}

	}
}

func TestSubscriberLength(t *testing.T) {
	s := Subscribers{}
	s.subscribers = map[string]map[string]bool{"test": map[string]bool{"test1": true}}
	if len(s.subscribers) != s.Length() {
		t.Error("Length Not Correct")
	}
}

func TestAddSubscriber(t *testing.T) {
	s := NewSubscribers()
	s.AddSubscriber("testClient1", "testPath")
	s.AddSubscriber("testClient2", "testPath")
	s.AddSubscriber("testClient3", "testPath2")
	if _, ok := s.subscribers["testPath"]; !ok {
		t.Error("Error Creating Path")
	} else if _, ok := s.subscribers["testPath"]["testClient1"]; !ok {
		t.Error("Error Setting testclient1")
	} else if _, ok := s.subscribers["testPath"]["testClient2"]; !ok {
		t.Error("Error Setting testclient2")
	} else if _, ok := s.subscribers["testPath2"]; !ok {
		t.Error("Error Setting testpath2")
	} else if _, ok := s.subscribers["testPath2"]["testClient3"]; !ok {
		t.Error("Error Setting testclient3")
	}
}

func TestDelSubscriber(t *testing.T) {
	s := Subscribers{}
	s.subscribers = map[string]map[string]bool{
		"testpath":  map[string]bool{"testclient1": true, "testclient2": true},
		"testpath2": map[string]bool{"testclient3": true}}
	s.DelSubscriber("testclient2", "testpath")
	if _, ok := s.subscribers["testpath"]["testclient2"]; ok {
		t.Error("client not deleted")
	} else if _, ok := s.subscribers["testpath"]["testclient1"]; !ok {
		t.Error("deleted wrong client")
	} else if _, ok := s.subscribers["testpath2"]["testclient3"]; !ok {
		t.Error("deleted wrong client")
	}
}

func TestDelSubscrberAll(t *testing.T) {
	s := Subscribers{}
	s.subscribers = map[string]map[string]bool{
		"testpath":  map[string]bool{"testclient1": true, "testclient2": true},
		"testpath2": map[string]bool{"testclient1": true}}

	s.DelSubscriberAll("testclient1")
	if _, ok := s.subscribers["testpath"]["testclient1"]; ok {
		t.Error("client not deleted")
	} else if _, ok := s.subscribers["testpath"]["testclient2"]; !ok {
		t.Error("deleted wrong client")
	} else if _, ok := s.subscribers["testpath2"]; ok {
		t.Error("didn't delte empty path")
	}
}

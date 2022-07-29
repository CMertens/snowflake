package snowflake

import (
	"fmt"
	"testing"
)

func TestSnowflakeConversion(t *testing.T) {
	s1 := NewSemanticSnowflake(2856524282194824821)
	sf := s1.ToSnowflake()
	s2 := NewSemanticSnowflake(sf)

	if s1.ID != s2.ID {
		t.Errorf("(1) ID %d converted to incorrect value %d!", s1.ID, s2.ID)
	}

	if s1.NodeID != s2.NodeID {
		t.Errorf("(2) System ID %d converted to incorrect value %d!", s1.NodeID, s2.NodeID)
	}

	if s1.TypeID != s2.TypeID {
		t.Errorf("(3) Class ID %d converted to incorrect value %d!", s1.TypeID, s2.TypeID)
	}

	s2.ID = 340524230265
	s2.NodeID = 50
	s2.TypeID = 100
	sf2 := s2.ToSnowflake()

	s3 := NewSemanticSnowflake(sf2)

	if s3.ID != s2.ID {
		t.Errorf("(4) ID %d converted to incorrect value %d!", s2.ID, s3.ID)
	}

	if s3.NodeID != s2.NodeID {
		t.Errorf("(5) System ID %d converted to incorrect value %d!", s2.NodeID, s3.NodeID)
	}

	if s3.TypeID != s2.TypeID {
		t.Errorf("(6) Class ID %d converted to incorrect value %d!", s2.TypeID, s3.TypeID)
	}

	fmt.Printf("OCID: %d\n", s3.GlobalTypeID)

}

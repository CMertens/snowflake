package snowflake

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"
)

// A "classic" snowflake is constructed as:
// gap_since_epoch_in_millis <<23 (41 bits) -- ID
// node_id << 10 (13 bits) -- Node ID
// nanosecond_deconflict (10 bits) -- Sub ID
//
// In a "semantic" snowflake, we assume the unique object-level
// key is in the "ID" field, the system or subsystem ID is
// in the "Node" field, and the class ID is in the "Sub ID"
// field. This allows us to mix semantic and non-semantic IDs, so long
// as the calling system knows which Nodes generate which kind of ID.
//
// There are up to 8191 System IDs available, and 1023 Class IDs for
// each System ID. The System and Class IDs form a class ID, which has a
// range of 1025 to 8,388,607 (excluding system ID 0). Passing a class ID
// to create a semantic snowflake will always return the system and
// class IDs.
// Snowflakes are always big-endian (network order).

const baseEpoch = int64(1611252000000)
const baseSeqIdBits = uint8(12)
const baseNodeBits = uint8(10)

type Snowflake int64

type NetSnowflake string

// Javascript has issues with int64s, so we expect IDs to be
// passed in as strings. This function unmarshals a string to
// an int64
func (sf *Snowflake) UnmarshalJSON(b []byte) error {
	var itm interface{}
	if err := json.Unmarshal(b, &itm); err != nil {
		return err
	}
	switch v := itm.(type) {
	case int:
		*sf = Snowflake(v)
	case float64:
		*sf = Snowflake(int(v))
	case int64:
		*sf = Snowflake(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		*sf = Snowflake(i)
	}
	return nil
}

func (sf Snowflake) MarshalJSON() ([]byte, error) {
	val := "\"" + strconv.FormatInt(int64(sf), 10) + "\""
	return []byte(val), nil
}

func FromString(id string) Snowflake {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return Snowflake(-1)
	}
	return Snowflake(i)
}

type SnowflakeNode struct {
	mutex      sync.Mutex
	sequence   int64
	epochBits  uint8
	nodeIdBits uint8
	seqIdBits  uint8
	epoch      time.Time

	seqStep  int64
	timeStep uint8
	nodeStep uint8
	time     int64
	nodeId   int64
}

func NewSnowflakeNode(shardId int) *SnowflakeNode {
	curTime := time.Now()
	var node SnowflakeNode = SnowflakeNode{
		sequence:   0,
		epochBits:  41,
		nodeIdBits: 10,
		seqIdBits:  baseSeqIdBits,
		nodeId:     int64(shardId),
		epoch:      curTime.Add(time.Unix(baseEpoch/1000, (baseEpoch%1000)*1000000).Sub(curTime)),
		seqStep:    -1 ^ (-1 << baseSeqIdBits),
		timeStep:   baseNodeBits + baseSeqIdBits,
		nodeStep:   baseNodeBits,
	}

	return &node
}

func (self *SnowflakeNode) Next() Snowflake {
	// Critical code -- prevent race conditions regarding the sequence
	self.mutex.Lock()
	now := time.Since(self.epoch).Nanoseconds() / 1000000
	if now == self.time {
		self.sequence = (self.sequence + 1) & self.seqStep
		if self.sequence == 0 {
			for now <= self.time {
				now = time.Since(self.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		self.sequence = 0
	}
	self.time = now
	seq := self.sequence
	self.mutex.Unlock()

	id := Snowflake(
		(now)<<self.timeStep |
			(self.nodeId << self.nodeStep) |
			(seq),
	)

	return id
}

func NewNetSnowflake(i int64) NetSnowflake {
	return NetSnowflake(strconv.FormatInt(i, 10))
}

func (s *NetSnowflake) Valid() bool {
	i, err := strconv.ParseInt(string(*s), 10, 64)
	if err != nil || i < 0 {
		return false
	}
	return true
}

func (s *NetSnowflake) ToID() int64 {
	i, err := strconv.ParseInt(string(*s), 10, 64)
	if err != nil {
		return -1
	}
	return i
}

type SemanticSnowflake struct {
	ID           int64
	NodeID       int64
	TypeID       int64
	GlobalTypeID int64
}

func NewSemanticSnowflake(flake Snowflake) SemanticSnowflake {
	// Snowflake format:
	// [TIMEST] [TIMEST] [TIMEST] [TIMEST] [TIMEST] [TSNODE] [NODECL] [CLASS ]
	// 00000000 00000000 00000000 00000000 00000000 01111111 11111122 22222222

	var id uint64 = uint64(flake)
	id = id >> 23

	var nodeid uint64 = uint64(flake)
	nodeid = nodeid << 41
	nodeid = nodeid >> 51

	var typeid uint64 = uint64(flake)
	typeid = typeid & ((1 << 10) - 1)

	var gtid uint64 = uint64(flake)
	gtid = gtid & ((1 << 23) - 1)

	return SemanticSnowflake{
		ID:           int64(id),
		NodeID:       int64(nodeid),
		TypeID:       int64(typeid),
		GlobalTypeID: int64(gtid),
	}
}

func (s *SemanticSnowflake) ToSnowflake() Snowflake {
	var i int64 = s.ID << 23
	i = i | (s.GetNodeID() << 10)
	i = i | (s.GetTypeID())
	return Snowflake(i)
}

func (s *SemanticSnowflake) ToNetSnowflake() NetSnowflake {
	return NewNetSnowflake(int64(s.ToSnowflake()))
}

func (s SemanticSnowflake) GetID() int64 {
	return s.ID
}

func (s SemanticSnowflake) GetNodeID() int64 {
	return int64(s.NodeID % 8192)
}

func (s SemanticSnowflake) GetTypeID() int64 {
	return int64(s.TypeID % 1024)
}

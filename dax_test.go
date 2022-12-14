package sabi

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var logs list.List
var WillFailToCreateFooDaxConn bool = false
var WillFailToCommitFooDaxConn bool = false

type /* error reason */ (
	InvalidDaxConn struct{}
)

func Clear() {
	isGlobalDaxSrcsFixed = false
	globalDaxSrcMap = make(map[string]DaxSrc)

	logs.Init()

	WillFailToCreateFooDaxConn = false
	WillFailToCommitFooDaxConn = false
}

type FooDaxConn struct {
	Label string
}

func (conn *FooDaxConn) Commit() Err {
	if WillFailToCommitFooDaxConn {
		return ErrBy(InvalidDaxConn{})
	}
	logs.PushBack("FooDaxConn#Commit")
	return Ok()
}

func (conn *FooDaxConn) Rollback() {
	logs.PushBack("FooDaxConn#Rollback")
}

func (conn *FooDaxConn) Close() {
	logs.PushBack("FooDaxConn#Close")
}

type FooDaxSrc struct {
	Label string
}

func (ds FooDaxSrc) CreateDaxConn() (DaxConn, Err) {
	if WillFailToCreateFooDaxConn {
		return nil, ErrBy(InvalidDaxConn{})
	}
	return &FooDaxConn{Label: ds.Label}, Ok()
}

type BarDaxConn struct {
	Label string
	store map[string]string
}

func (conn *BarDaxConn) Commit() Err {
	logs.PushBack("BarDaxConn#Commit")
	return Ok()
}

func (conn *BarDaxConn) Rollback() {
	logs.PushBack("BarDaxConn#Rollback")
}

func (conn *BarDaxConn) Close() {
	logs.PushBack("BarDaxConn#Close")
}

func (conn *BarDaxConn) Store(name, value string) {
	conn.store[name] = value
}

type BarDaxSrc struct {
	Label string
	Store map[string]string
}

func (ds BarDaxSrc) CreateDaxConn() (DaxConn, Err) {
	return &BarDaxConn{Label: ds.Label, store: ds.Store}, Ok()
}

func TestAddGlobalDaxSrc(t *testing.T) {
	Clear()
	defer Clear()

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 0)

	AddGlobalDaxSrc("foo", FooDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 2)
}

func TestFixGlobalDaxSrcs(t *testing.T) {
	Clear()
	defer Clear()

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 0)

	AddGlobalDaxSrc("foo", FooDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)

	FixGlobalDaxSrcs()

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)

	isGlobalDaxSrcsFixed = false

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 2)
}

func TestDaxBase_AddLocalDaxSrc(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(base.localDaxSrcMap), 0)
	assert.Equal(t, len(base.daxConnMap), 0)

	base.AddLocalDaxSrc("foo", FooDaxSrc{})

	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(base.localDaxSrcMap), 1)
	assert.Equal(t, len(base.daxConnMap), 0)

	base.AddLocalDaxSrc("bar", &BarDaxSrc{})

	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(base.localDaxSrcMap), 2)
	assert.Equal(t, len(base.daxConnMap), 0)
}

func TestDaxBase_begin(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 0)
	assert.Equal(t, len(base.localDaxSrcMap), 0)
	assert.Equal(t, len(base.daxConnMap), 0)

	AddGlobalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("foo", FooDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 1)
	assert.Equal(t, len(base.daxConnMap), 0)

	base.begin()

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.True(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 1)
	assert.Equal(t, len(base.daxConnMap), 0)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.True(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 1)
	assert.Equal(t, len(base.daxConnMap), 0)

	base.isLocalDaxSrcsFixed = false

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 1)
	assert.Equal(t, len(base.daxConnMap), 0)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})

	assert.True(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 2)
	assert.Equal(t, len(base.daxConnMap), 0)

	isGlobalDaxSrcsFixed = false

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 1)
	assert.Equal(t, len(base.localDaxSrcMap), 2)
	assert.Equal(t, len(base.daxConnMap), 0)

	AddGlobalDaxSrc("bar", &BarDaxSrc{})

	assert.False(t, isGlobalDaxSrcsFixed)
	assert.False(t, base.isLocalDaxSrcsFixed)
	assert.Equal(t, len(globalDaxSrcMap), 2)
	assert.Equal(t, len(base.localDaxSrcMap), 2)
	assert.Equal(t, len(base.daxConnMap), 0)
}

func TestDaxBase_GetDaxConn_withLocalDaxSrc(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	conn, err := base.GetDaxConn("foo")
	assert.Nil(t, conn)
	switch err.Reason().(type) {
	case DaxSrcIsNotFound:
		assert.Equal(t, err.Get("Name"), "foo")
	default:
		assert.Fail(t, err.Error())
	}

	base.AddLocalDaxSrc("foo", FooDaxSrc{})

	conn, err = base.GetDaxConn("foo")
	assert.NotNil(t, conn)
	assert.True(t, err.IsOk())

	var conn2 DaxConn
	conn2, err = base.GetDaxConn("foo")
	assert.Equal(t, conn2, conn)
	assert.True(t, err.IsOk())
}

func TestDaxBase_GetDaxConn_withGlobalDaxSrc(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	conn, err := base.GetDaxConn("foo")
	assert.Nil(t, conn)
	switch err.Reason().(type) {
	case DaxSrcIsNotFound:
		assert.Equal(t, err.Get("Name"), "foo")
	default:
		assert.Fail(t, err.Error())
	}

	AddGlobalDaxSrc("foo", FooDaxSrc{})

	conn, err = base.GetDaxConn("foo")
	assert.NotNil(t, conn)
	assert.True(t, err.IsOk())

	var conn2 DaxConn
	conn2, err = base.GetDaxConn("foo")
	assert.Equal(t, conn2, conn)
	assert.True(t, err.IsOk())
}

func TestDaxBase_GetDaxConn_localDsIsTakenPriorityOfGlobalDs(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	conn, err := base.GetDaxConn("foo")
	assert.Nil(t, conn)
	switch err.Reason().(type) {
	case DaxSrcIsNotFound:
		assert.Equal(t, err.Get("Name"), "foo")
	default:
		assert.Fail(t, err.Error())
	}

	AddGlobalDaxSrc("foo", FooDaxSrc{Label: "global"})
	FixGlobalDaxSrcs()

	base.AddLocalDaxSrc("foo", FooDaxSrc{Label: "local"})

	conn, err = base.GetDaxConn("foo")
	assert.Equal(t, conn.(*FooDaxConn).Label, "local")
	assert.True(t, err.IsOk())
}

func TestDaxBase_GetDaxConn_failToCreateDaxConn(t *testing.T) {
	Clear()
	defer Clear()

	WillFailToCreateFooDaxConn = true
	defer func() { WillFailToCreateFooDaxConn = false }()

	base := NewDaxBase()
	base.AddLocalDaxSrc("foo", FooDaxSrc{})

	conn, err := base.GetDaxConn("foo")
	assert.Nil(t, conn)
	switch err.Reason().(type) {
	case FailToCreateDaxConn:
		assert.Equal(t, err.Get("Name"), "foo")
		switch err.Cause().(Err).Reason().(type) {
		case InvalidDaxConn:
		default:
			assert.Fail(t, err.Error())
		}
	default:
		assert.Fail(t, err.Error())
	}
}

func TestDaxBase_commit(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	base.AddLocalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})
	base.begin()

	fooConn, fooErr := base.GetDaxConn("foo")
	assert.NotNil(t, fooConn)
	assert.True(t, fooErr.IsOk())

	barConn, barErr := base.GetDaxConn("bar")
	assert.NotNil(t, barConn)
	assert.True(t, barErr.IsOk())

	err := base.commit()
	assert.True(t, err.IsOk())

	assert.Equal(t, logs.Len(), 2)
	if logs.Front().Value == "FooDaxConn#Commit" {
		assert.Equal(t, logs.Front().Value, "FooDaxConn#Commit")
		assert.Equal(t, logs.Back().Value, "BarDaxConn#Commit")
	} else {
		assert.Equal(t, logs.Front().Value, "BarDaxConn#Commit")
		assert.Equal(t, logs.Back().Value, "FooDaxConn#Commit")
	}
}

func TestDaxBase_commit_failed(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	base.AddLocalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})

	base.begin()

	fooConn, fooErr := base.GetDaxConn("foo")
	assert.NotNil(t, fooConn)
	assert.True(t, fooErr.IsOk())

	barConn, barErr := base.GetDaxConn("bar")
	assert.NotNil(t, barConn)
	assert.True(t, barErr.IsOk())

	WillFailToCommitFooDaxConn = true

	err := base.commit()
	assert.False(t, err.IsOk())
	switch err.Reason().(type) {
	case FailToCommitDaxConn:
		m := err.Get("Errors").(map[string]Err)
		assert.Equal(t, m["foo"].ReasonName(), "InvalidDaxConn")
	default:
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, logs.Len(), 1)
	assert.Equal(t, logs.Back().Value, "BarDaxConn#Commit")
}

func TestDaxBase_rollback(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	base.AddLocalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})
	base.begin()

	fooConn, fooErr := base.GetDaxConn("foo")
	assert.NotNil(t, fooConn)
	assert.True(t, fooErr.IsOk())

	barConn, barErr := base.GetDaxConn("bar")
	assert.NotNil(t, barConn)
	assert.True(t, barErr.IsOk())

	base.rollback()

	assert.Equal(t, logs.Len(), 2)
	if logs.Front().Value == "FooDaxConn#Rollback" {
		assert.Equal(t, logs.Front().Value, "FooDaxConn#Rollback")
		assert.Equal(t, logs.Back().Value, "BarDaxConn#Rollback")
	} else {
		assert.Equal(t, logs.Front().Value, "BarDaxConn#Rollback")
		assert.Equal(t, logs.Back().Value, "FooDaxConn#Rollback")
	}
}

func TestDaxBase_close(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()

	base.AddLocalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})
	base.begin()

	fooConn, fooErr := base.GetDaxConn("foo")
	assert.NotNil(t, fooConn)
	assert.True(t, fooErr.IsOk())

	barConn, barErr := base.GetDaxConn("bar")
	assert.NotNil(t, barConn)
	assert.True(t, barErr.IsOk())

	base.close()

	assert.Equal(t, logs.Len(), 2)
	if logs.Front().Value == "FooDaxConn#Close" {
		assert.Equal(t, logs.Front().Value, "FooDaxConn#Close")
		assert.Equal(t, logs.Back().Value, "BarDaxConn#Close")
	} else {
		assert.Equal(t, logs.Front().Value, "BarDaxConn#Close")
		assert.Equal(t, logs.Back().Value, "FooDaxConn#Close")
	}
}

type FooDax struct {
	Dax
}

func NewFooDax(dax Dax) FooDax {
	return FooDax{Dax: dax}
}

func (dax FooDax) GetFooDaxConn(name string) (*FooDaxConn, Err) {
	conn, err := dax.GetDaxConn(name)
	if !err.IsOk() {
		return nil, err
	}
	return conn.(*FooDaxConn), Ok()
}

type BarDax struct {
	Dax
}

func NewBarDax(dax Dax) BarDax {
	return BarDax{Dax: dax}
}

func (dax BarDax) GetBarDaxConn(name string) (*BarDaxConn, Err) {
	conn, err := dax.GetDaxConn(name)
	if !err.IsOk() {
		return nil, err
	}
	return conn.(*BarDaxConn), Ok()
}

func TestDax_GetXxxConn(t *testing.T) {
	Clear()
	defer Clear()

	base := NewDaxBase()
	base.AddLocalDaxSrc("foo", FooDaxSrc{})
	base.AddLocalDaxSrc("bar", &BarDaxSrc{})

	base.begin()

	fooDax := NewFooDax(base)
	fooConn, fooErr := fooDax.GetFooDaxConn("foo")
	assert.True(t, fooErr.IsOk())
	assert.Equal(t, reflect.TypeOf(fooConn).String(), "*sabi.FooDaxConn")

	barDax := NewBarDax(base)
	barConn, barErr := barDax.GetBarDaxConn("bar")
	assert.True(t, barErr.IsOk())
	assert.Equal(t, reflect.TypeOf(barConn).String(), "*sabi.BarDaxConn")
}

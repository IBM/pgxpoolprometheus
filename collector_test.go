package pgxpoolprometheus

/**
 * (C) Copyright IBM Corp. 2021.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStater struct {
	mock.Mock
}

func (m *mockStater) Stat() pgxStat {
	return m.Called().Get(0).(pgxStat)
}

var (
	_ pgxStat = (*pgxStatMock)(nil)
	_ pgxStat = (*noOpStat)(nil)
)

type pgxStatMock struct {
	mock.Mock
}

func (m *pgxStatMock) AcquireCount() int64 {
	return m.Called().Get(0).(int64)
}
func (m *pgxStatMock) AcquireDuration() time.Duration {
	return m.Called().Get(0).(time.Duration)
}
func (m *pgxStatMock) AcquiredConns() int32 {
	return m.Called().Get(0).(int32)
}
func (m *pgxStatMock) CanceledAcquireCount() int64 {
	return m.Called().Get(0).(int64)
}
func (m *pgxStatMock) ConstructingConns() int32 {
	return m.Called().Get(0).(int32)
}
func (m *pgxStatMock) EmptyAcquireCount() int64 {
	return m.Called().Get(0).(int64)
}
func (m *pgxStatMock) IdleConns() int32 {
	return m.Called().Get(0).(int32)
}
func (m *pgxStatMock) MaxConns() int32 {
	return m.Called().Get(0).(int32)
}
func (m *pgxStatMock) TotalConns() int32 {
	return m.Called().Get(0).(int32)
}
func (m *pgxStatMock) NewConnsCount() int64 {
	return m.Called().Get(0).(int64)
}
func (m *pgxStatMock) MaxLifetimeDestroyCount() int64 {
	return m.Called().Get(0).(int64)
}
func (m *pgxStatMock) MaxIdleDestroyCount() int64 {
	return m.Called().Get(0).(int64)
}

type noOpStat struct{}

func (m noOpStat) AcquireCount() int64 {
	return 0
}
func (m noOpStat) AcquireDuration() time.Duration {
	return time.Second * 0
}
func (m noOpStat) AcquiredConns() int32 {
	return 0
}
func (m noOpStat) CanceledAcquireCount() int64 {
	return 0
}
func (m noOpStat) ConstructingConns() int32 {
	return 0
}
func (m noOpStat) EmptyAcquireCount() int64 {
	return 0
}
func (m noOpStat) IdleConns() int32 {
	return 0
}
func (m noOpStat) MaxConns() int32 {
	return 0
}
func (m noOpStat) TotalConns() int32 {
	return 0
}
func (m noOpStat) NewConnsCount() int64 {
	return 0
}
func (m noOpStat) MaxLifetimeDestroyCount() int64 {
	return 0
}
func (m noOpStat) MaxIdleDestroyCount() int64 {
	return 0
}

func TestDescribeDescribesAllAvailableStats(t *testing.T) {
	labelName := "testLabel"
	labelValue := "testLabelValue"
	labels := map[string]string{labelName: labelValue}
	expectedDescriptorCount := 12
	timeout := time.After(time.Second * 5)
	stater := &mockStater{}
	stater.On("Stat").Return(noOpStat{})
	statFn := func() pgxStat { return stater.Stat() }
	testObject := newCollector(statFn, labels)

	ch := make(chan *prometheus.Desc)
	go testObject.Describe(ch)

	expectedDescriptorCountRemaining := expectedDescriptorCount
	uniqueDescriptors := make(map[string]struct{})
	for {
		if expectedDescriptorCountRemaining == 0 {
			break
		}
		select {
		case desc := <-ch:
			assert.Contains(t, desc.String(), labelName)
			assert.Contains(t, desc.String(), labelValue)
			uniqueDescriptors[desc.String()] = struct{}{}
			expectedDescriptorCountRemaining--
		case <-timeout:
			t.Fatalf("Test timed out while there were still %d descriptors expected", expectedDescriptorCountRemaining)
		}
	}
	assert.Equal(t, 0, expectedDescriptorCountRemaining)
	assert.Len(t, uniqueDescriptors, expectedDescriptorCount)
}

func TestCollectCollectsAllAvailableStats(t *testing.T) {
	expectedAcquireCount := float64(1)
	expectedAcquireDuration := float64(2e+09)
	expectedAcquiredConns := float64(3)
	expectedCanceledAcquireCount := float64(4)
	expectedConstructingConns := float64(5)
	expectedEmptyAcquireCount := float64(6)
	expectedIdleConns := float64(7)
	expectedMaxConns := float64(8)
	expectedTotalConns := float64(9)
	expectedNewConnsCount := float64(10)
	expectedMaxLifetimeDestroyCount := float64(11)
	expectedMaxIdleDestroyCount := float64(12)

	mockStats := &pgxStatMock{}
	mockStats.On("AcquireCount").Return(int64(1))
	mockStats.On("AcquireDuration").Return(time.Second * 2)
	mockStats.On("AcquiredConns").Return(int32(3))
	mockStats.On("CanceledAcquireCount").Return(int64(4))
	mockStats.On("ConstructingConns").Return(int32(5))
	mockStats.On("EmptyAcquireCount").Return(int64(6))
	mockStats.On("IdleConns").Return(int32(7))
	mockStats.On("MaxConns").Return(int32(8))
	mockStats.On("TotalConns").Return(int32(9))
	mockStats.On("NewConnsCount").Return(int64(10))
	mockStats.On("MaxLifetimeDestroyCount").Return(int64(11))
	mockStats.On("MaxIdleDestroyCount").Return(int64(12))
	expectedMetricCount := 12
	timeout := time.After(time.Second * 5)
	stater := &mockStater{}
	stater.On("Stat").Return(mockStats)
	staterfn := func() pgxStat { return stater.Stat() }
	testObject := newCollector(staterfn, nil)

	ch := make(chan prometheus.Metric)
	go testObject.Collect(ch)

	expectedMetricCountRemaining := expectedMetricCount
	for {
		if expectedMetricCountRemaining == 0 {
			break
		}
		select {
		case metric := <-ch:
			pb := &dto.Metric{}
			metric.Write(pb)
			description := metric.Desc().String()
			switch {
			case strings.Contains(description, "pgxpool_acquire_count"):
				assert.Equal(t, expectedAcquireCount, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_acquire_duration_ns"):
				assert.Equal(t, expectedAcquireDuration, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_acquired_conns"):
				assert.Equal(t, expectedAcquiredConns, *pb.GetGauge().Value)
			case strings.Contains(description, "pgxpool_canceled_acquire_count"):
				assert.Equal(t, expectedCanceledAcquireCount, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_constructing_conns"):
				assert.Equal(t, expectedConstructingConns, *pb.GetGauge().Value)
			case strings.Contains(description, "pgxpool_empty_acquire"):
				assert.Equal(t, expectedEmptyAcquireCount, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_idle_conns"):
				assert.Equal(t, expectedIdleConns, *pb.GetGauge().Value)
			case strings.Contains(description, "pgxpool_max_conns"):
				assert.Equal(t, expectedMaxConns, *pb.GetGauge().Value)
			case strings.Contains(description, "pgxpool_total_conns"):
				assert.Equal(t, expectedTotalConns, *pb.GetGauge().Value)
			case strings.Contains(description, "pgxpool_new_conns_count"):
				assert.Equal(t, expectedNewConnsCount, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_max_lifetime_destroy_count"):
				assert.Equal(t, expectedMaxLifetimeDestroyCount, *pb.GetCounter().Value)
			case strings.Contains(description, "pgxpool_max_idle_destroy_count"):
				assert.Equal(t, expectedMaxIdleDestroyCount, *pb.GetCounter().Value)
			default:
				t.Errorf("Unexpected description: %s", description)
			}
			expectedMetricCountRemaining--
		case <-timeout:
			t.Fatalf("Test timed out while there were still %d descriptors expected", expectedMetricCountRemaining)
		}
	}
	assert.Equal(t, 0, expectedMetricCountRemaining)
}

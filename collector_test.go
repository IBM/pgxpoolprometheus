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

func (m *mockStater) stat() stat {
	return m.Called().Get(0).(stat)
}

type mockStat struct {
	mock.Mock
}

func (m *mockStat) acquireCount() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) acquireDuration() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) acquiredConns() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) canceledAcquireCount() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) constructingConns() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) emptyAcquireCount() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) idleConns() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) maxConns() float64 {
	return m.Called().Get(0).(float64)
}
func (m *mockStat) totalConns() float64 {
	return m.Called().Get(0).(float64)
}

type noOpStat struct{}

func (m noOpStat) acquireCount() float64 {
	return 0
}
func (m noOpStat) acquireDuration() float64 {
	return 0
}
func (m noOpStat) acquiredConns() float64 {
	return 0
}
func (m noOpStat) canceledAcquireCount() float64 {
	return 0
}
func (m noOpStat) constructingConns() float64 {
	return 0
}
func (m noOpStat) emptyAcquireCount() float64 {
	return 0
}
func (m noOpStat) idleConns() float64 {
	return 0
}
func (m noOpStat) maxConns() float64 {
	return 0
}
func (m noOpStat) totalConns() float64 {
	return 0
}

func TestDescribeDescribesAllAvailableStats(t *testing.T) {
	labelName := "testLabel"
	labelValue := "testLabelValue"
	labels := map[string]string{labelName: labelValue}
	expectedDescriptorCount := 9
	timeout := time.After(time.Second * 5)
	stater := &mockStater{}
	stater.On("stat").Return(noOpStat{})
	testObject := newCollector(stater, labels)

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
	expectedAcquireDuration := float64(2)
	expectedAcquiredConns := float64(3)
	expectedCanceledAcquireCount := float64(4)
	expectedConstructingConns := float64(5)
	expectedEmptyAcquireCount := float64(6)
	expectedIdleConns := float64(7)
	expectedMaxConns := float64(8)
	expectedTotalConns := float64(9)
	mockStats := &mockStat{}
	mockStats.On("acquireCount").Return(expectedAcquireCount)
	mockStats.On("acquireDuration").Return(expectedAcquireDuration)
	mockStats.On("acquiredConns").Return(expectedAcquiredConns)
	mockStats.On("canceledAcquireCount").Return(expectedCanceledAcquireCount)
	mockStats.On("constructingConns").Return(expectedConstructingConns)
	mockStats.On("emptyAcquireCount").Return(expectedEmptyAcquireCount)
	mockStats.On("idleConns").Return(expectedIdleConns)
	mockStats.On("maxConns").Return(expectedMaxConns)
	mockStats.On("totalConns").Return(expectedTotalConns)
	expectedMetricCount := 9
	timeout := time.After(time.Second * 5)
	stater := &mockStater{}
	stater.On("stat").Return(mockStats)
	testObject := newCollector(stater, nil)

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

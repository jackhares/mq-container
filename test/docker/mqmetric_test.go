/*
© Copyright IBM Corporation 2018

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
)

func TestGoldenPathMetric(t *testing.T) {
	t.Parallel()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	// hostname := getIPAddress(t, cli, id)
	port := getMetricPort(t, cli, id)
	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)

	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)
	time.Sleep(15 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned but had none...")
		t.Fail()
	}
	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestMetricNames(t *testing.T) {
	t.Parallel()
	approvedSuffixes := []string{"bytes", "seconds", "percentage", "count", "total"}
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	port := getMetricPort(t, cli, id)
	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)
	// Call once as mq_prometheus 'ignores' the first call
	getMetrics(t, port)
	time.Sleep(15 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned but had none...")
		t.Fail()
	}

	okMetrics := []string{}
	badMetrics := []string{}

	for _, metric := range metrics {
		ok := false
		for _, e := range approvedSuffixes {
			if strings.HasSuffix(metric.Key, e) {
				ok = true
				break
			}
		}

		if !ok {
			t.Logf("Metric '%s' does not have an approved suffix", metric.Key)
			badMetrics = append(badMetrics, metric.Key)
			t.Fail()
		} else {
			okMetrics = append(okMetrics, metric.Key)
		}
	}

	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestMetricLabels(t *testing.T) {
	t.Parallel()
	requiredLabels := []string{"qmgr"}
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	port := getMetricPort(t, cli, id)
	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)
	// Call once as mq_prometheus 'ignores' the first call
	getMetrics(t, port)
	time.Sleep(15 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Error("Expected some metrics to be returned but had none")
	}

	for _, metric := range metrics {
		found := false
		for key := range metric.Labels {
			for _, e := range requiredLabels {
				if key == e {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			t.Errorf("Metric '%s' with labels %s does not have one or more required labels - %s", metric.Key, metric.Labels, requiredLabels)
		}
	}
	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestRapidFirePrometheus(t *testing.T) {
	t.Parallel()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	port := getMetricPort(t, cli, id)
	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)
	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)
	// Rapid fire it then check we're still happy
	for i := 0; i < 30; i++ {
		getMetrics(t, port)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(11 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Error("Expected some metrics to be returned but had none")
	}
	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestSlowPrometheus(t *testing.T) {
	t.Parallel()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	port := getMetricPort(t, cli, id)
	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)
	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)
	// Send a request twice over a long period and check we're still happy
	for i := 0; i < 2; i++ {
		time.Sleep(30 * time.Second)
		metrics := getMetrics(t, port)
		if len(metrics) <= 0 {
			t.Log("Expected some metrics to be returned but had none")
			t.Fail()
		}

	}
	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestContainerRestart(t *testing.T) {
	t.Parallel()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)
	port := getMetricPort(t, cli, id)

	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)

	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)

	time.Sleep(15 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned before the restart but had none...")
		t.FailNow()
	}

	// Stop the container cleanly
	stopContainer(t, cli, id)
	// Start the container cleanly
	startContainer(t, cli, id)

	port = getMetricPort(t, cli, id)

	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)

	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)

	time.Sleep(15 * time.Second)
	metrics = getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned before the restart but had none...")
		t.Fail()
	}
}

func TestQMRestart(t *testing.T) {
	t.Parallel()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	id := runContainerWithPorts(t, cli, metricsContainerConfig(), []int{defaultMetricPort})
	defer cleanContainer(t, cli, id)

	port := getMetricPort(t, cli, id)

	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)

	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)

	time.Sleep(15 * time.Second)
	metrics := getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned before the restart but had none...")
		t.FailNow()
	}

	// Restart just the QM (to simulate a lost connection)
	t.Log("Stopping queue manager\n")
	rc, out := execContainer(t, cli, id, "mqm", []string{"endmqm", "-w", defaultMetricQMName})
	if rc != 0 {
		t.Logf("Failed to stop the queue manager. rc=%d, err=%s", rc, out)
		t.FailNow()
	}
	t.Log("starting queue manager\n")
	rc, out = execContainer(t, cli, id, "mqm", []string{"strmqm", defaultMetricQMName})
	if rc != 0 {
		t.Logf("Failed to start the queue manager. rc=%d, err=%s", rc, out)
		t.FailNow()
	}

	// Wait for the queue manager to come back up
	time.Sleep(10 * time.Second)

	// Now the container is ready we prod the prometheus endpoint until it's up.
	waitForMetricReady(t, port)

	// Call once as mq_prometheus 'ignores' the first call and will not return any metrics
	getMetrics(t, port)

	time.Sleep(15 * time.Second)
	metrics = getMetrics(t, port)
	if len(metrics) <= 0 {
		t.Log("Expected some metrics to be returned before the restart but had none...")
		t.FailNow()
	}

	// Stop the container cleanly
	stopContainer(t, cli, id)
}

func TestValidValues(t *testing.T) {

}

func TestChangingValues(t *testing.T) {

}
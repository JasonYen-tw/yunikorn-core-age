/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package objects

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/apache/yunikorn-core/pkg/common"
	"github.com/apache/yunikorn-core/pkg/events"
	"github.com/apache/yunikorn-core/pkg/metrics"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

func TestAcceptStateTransition(t *testing.T) {
	// Accept only from new
	app := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, app.CurrentState(), New.String())

	// new to accepted
	err := app.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted")
	assert.Equal(t, app.CurrentState(), Accepted.String())

	// accepted to rejected: error expected
	err = app.HandleApplicationEvent(RejectApplication)
	assert.Assert(t, err != nil, "error expected accepted to rejected")
	assert.Equal(t, app.CurrentState(), Accepted.String())

	// accepted to failed
	err = app.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected accepted to failing")
	err = common.WaitForCondition(10*time.Microsecond, time.Millisecond*100, app.IsFailing)
	assert.NilError(t, err, "App should be in Failing state")
}

func TestRejectStateTransition(t *testing.T) {
	// Reject only from new
	app := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, app.CurrentState(), New.String())

	// new to rejected
	err := app.HandleApplicationEvent(RejectApplication)
	assert.NilError(t, err, "no error expected new to rejected")
	assert.Equal(t, app.CurrentState(), Rejected.String())

	// app already rejected: error expected
	err = app.HandleApplicationEvent(RejectApplication)
	assert.Assert(t, err != nil, "error expected rejected to rejected")
	assert.Equal(t, app.CurrentState(), Rejected.String())

	// rejected to failing: error expected
	err = app.HandleApplicationEvent(FailApplication)
	assert.Assert(t, err != nil, "error expected rejected to failing")
	assert.Equal(t, app.CurrentState(), Rejected.String())
}

func TestRunStateTransition(t *testing.T) {
	// run from new
	appInfo := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, appInfo.CurrentState(), New.String())
	err := appInfo.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted (run test)")
	err = appInfo.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected accepted to running (run test)")
	assert.Equal(t, appInfo.CurrentState(), Running.String())

	// run app: same state is allowed for running
	err = appInfo.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected running to running")
	assert.Equal(t, appInfo.CurrentState(), Running.String())

	// running to rejected: error expected
	err = appInfo.HandleApplicationEvent(RejectApplication)
	assert.Assert(t, err != nil, "error expected running to rejected")
	assert.Equal(t, appInfo.CurrentState(), Running.String())

	// run to failing
	err = appInfo.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected running to failing")
	err = common.WaitForCondition(10*time.Microsecond, time.Millisecond*100, appInfo.IsFailing)
	assert.NilError(t, err, "App should be in Failing state")

	// run fails from failing
	err = appInfo.HandleApplicationEvent(RunApplication)
	assert.Assert(t, err != nil, "error expected failing to running")
	assert.Equal(t, appInfo.CurrentState(), Failing.String())
}

func TestCompletedStateTransition(t *testing.T) {
	// complete only from completing
	appInfo1 := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, appInfo1.CurrentState(), New.String())
	err := appInfo1.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted (completed test)")
	err = appInfo1.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected accepted to running (completed test)")
	err = appInfo1.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected running to running (completed test)")
	assert.Equal(t, appInfo1.CurrentState(), Running.String())
	// completed from run through completing
	err = appInfo1.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected running to completing")
	assert.Equal(t, appInfo1.CurrentState(), Completing.String())
	err = appInfo1.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected completing to completed")
	assert.Equal(t, appInfo1.CurrentState(), Completed.String())

	// complete from run through completing
	appInfo2 := newApplication("app-00002", "default", "root.a")
	assert.Equal(t, appInfo2.CurrentState(), New.String())
	err = appInfo2.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted (completed test)")
	err = appInfo2.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected accepted to completing (completed test)")
	assert.Equal(t, appInfo2.CurrentState(), Completing.String())
	// completed from completing
	err = appInfo2.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected completing to completed")
	assert.Equal(t, appInfo2.CurrentState(), Completed.String())

	// completed to rejected: error expected
	err = appInfo2.HandleApplicationEvent(RejectApplication)
	assert.Assert(t, err != nil, "error expected completed to rejected")
	assert.Equal(t, appInfo2.CurrentState(), Completed.String())

	// completed to failing: error expected
	err = appInfo2.HandleApplicationEvent(FailApplication)
	assert.Assert(t, err != nil, "error expected completed to failing")
	assert.Equal(t, appInfo2.CurrentState(), Completed.String())

	// completed fails from all completing
	appInfo3 := newApplication("app-00003", "default", "root.a")
	assert.Equal(t, appInfo3.CurrentState(), New.String())
	err = appInfo3.HandleApplicationEvent(CompleteApplication)
	assert.Assert(t, err != nil, "error expected new to completed")
	assert.Equal(t, appInfo3.CurrentState(), New.String())
}

func TestCompletingStateTransition(t *testing.T) {
	appInfo1 := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, appInfo1.CurrentState(), New.String())
	err := appInfo1.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted (completing test)")

	// accepted to completing and back again
	err = appInfo1.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected accepted to completing")
	assert.Equal(t, appInfo1.CurrentState(), Completing.String())

	// running to completing
	appInfo2 := newApplication("app-00002", "default", "root.a")
	assert.Equal(t, appInfo2.CurrentState(), New.String())
	err = appInfo2.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted (completing test)")
	err = appInfo2.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected accepted to running")
	assert.Equal(t, appInfo2.CurrentState(), Running.String())
	err = appInfo2.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected running to completing")
	assert.Equal(t, appInfo2.CurrentState(), Completing.String())

	// completing to run and back again
	err = appInfo2.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected completing to running (completing test)")
	err = appInfo2.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected running to completing")
	assert.Equal(t, appInfo2.CurrentState(), Completing.String())
}

func TestFailedStateTransition(t *testing.T) {
	// failing from all but rejected & completed
	appInfo := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, appInfo.CurrentState(), New.String())

	// new to failing
	err := appInfo.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected new to failing")
	err = common.WaitForCondition(10*time.Microsecond, time.Millisecond*100, appInfo.IsFailing)
	assert.NilError(t, err, "App should be in Failing state")

	// failing to failed
	err = appInfo.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected failing to failed")
	assert.Assert(t, appInfo.IsFailed(), "App should be in Failed state")

	// failed to rejected: error expected
	err = appInfo.HandleApplicationEvent(RejectApplication)
	assert.Assert(t, err != nil, "error expected failing to rejected")
	err = common.WaitForCondition(10*time.Microsecond, time.Millisecond*100, appInfo.IsFailed)
	assert.NilError(t, err, "App should be in Failed state")
}

func TestAppStateTransitionEvents(t *testing.T) {
	events.Init()
	eventSystem := events.GetEventSystem().(*events.EventSystemImpl) //nolint:errcheck
	eventSystem.StartServiceWithPublisher(false)
	// Accept only from new
	appInfo := newApplication("app-00001", "default", "root.a")
	assert.Equal(t, appInfo.CurrentState(), New.String())

	// new to accepted
	err := appInfo.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected new to accepted")
	assert.Equal(t, appInfo.CurrentState(), Accepted.String())

	// accepted to completing
	err = appInfo.HandleApplicationEvent(CompleteApplication)
	assert.NilError(t, err, "no error expected accepted to completing")
	assert.Equal(t, appInfo.CurrentState(), Completing.String())

	// completing to run
	err = appInfo.HandleApplicationEvent(RunApplication)
	assert.NilError(t, err, "no error expected completing to running (completing test)")

	// run to failing
	err = appInfo.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected new to failing")
	err = common.WaitForCondition(10*time.Microsecond, time.Millisecond*100, appInfo.IsFailing)
	assert.NilError(t, err, "App should be in Failing state")

	// failing to failed
	err = appInfo.HandleApplicationEvent(FailApplication)
	assert.NilError(t, err, "no error expected failing to failed")
	assert.Assert(t, appInfo.IsFailed(), "App should be in Failed state")

	// failed to expired
	err = appInfo.HandleApplicationEvent(ExpireApplication)
	assert.NilError(t, err, "no error expected failed to expired")
	assert.Assert(t, appInfo.IsExpired(), "App should be in Expired state")

	// accepted to resuming
	appInfo.stateMachine.SetState(Accepted.String())
	err = appInfo.HandleApplicationEvent(ResumeApplication)
	assert.NilError(t, err, "no error expected accepted to resuming")
	assert.Assert(t, appInfo.IsResuming(), "App should be in Resuming state")

	// Verify application events
	err = common.WaitForCondition(10*time.Millisecond, time.Second, func() bool {
		fmt.Printf("checking event length: %d\n", eventSystem.Store.CountStoredEvents())
		return eventSystem.Store.CountStoredEvents() == 8
	})
	assert.NilError(t, err, "the event should have been processed")
	records := eventSystem.Store.CollectEvents()
	if records == nil {
		t.Fatal("collecting eventChannel should return something")
	}
	assert.Equal(t, 8, len(records), "expecting 8 events")
	isNewApplicationEvent(t, appInfo, records[0])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_ACCEPTED, records[1])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_COMPLETING, records[2])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_RUNNING, records[3])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_FAILING, records[4])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_FAILED, records[5])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_EXPIRED, records[6])
	isStateChangeEvent(t, appInfo, si.EventRecord_APP_RESUMING, records[7])
}

// Test to verify metrics after applications state transition
// app-00001: New -> Resuming -> Accepted -> Running -> Completing-> Completed
// app-00002: New -> Accepted -> Running -> Completing -> Running -> Failing-> Failed
// app-00003: New -> Accepted -> Running -> Failing -> Failed
// app-00004: New -> Rejected
// Final queue metrics will be: 0 new, 0 running, 0 accepted, 1 completed, 2 failed and 1 rejected applications
// Final scheduler metrics will be: 4 new, 0 running, 3 accepted, 1 completed, 2 failed and 1 rejected applications
// Because the scheduler app submission state (new, accepted, rejected)is counter based, and it will not be decremented when the state changes
func TestAppStateTransitionMetrics(t *testing.T) { //nolint:funlen
	queue := createQueue(t, "metrics")
	metrics.GetSchedulerMetrics().Reset()
	metrics.GetQueueMetrics("root.metrics").Reset()
	// app-00001: New -> Resuming -> Accepted --> Running -> Completing-> Completed
	app := newApplication("app-00001", "default", "root.metrics")
	app.SetQueue(queue)
	assertState(t, app, nil, New.String())
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsNewMetrics(t, app, 1)

	// New -> Resuming
	err := app.HandleApplicationEvent(ResumeApplication)
	assertState(t, app, err, Resuming.String())
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsResumingMetrics(t, app, 1)
	assertQueueApplicationsNewMetrics(t, app, 0)
	// Resuming -> Accepted
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Accepted.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 1)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsNewMetrics(t, app, 0)
	// Accepted -> Running
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Running.String())
	assertTotalAppsRunningMetrics(t, 1)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertQueueRunningApps(t, app, 1)
	assertQueueApplicationsRunningMetrics(t, app, 1)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsNewMetrics(t, app, 0)
	// Running -> Running
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Running.String())
	assertTotalAppsRunningMetrics(t, 1)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertQueueRunningApps(t, app, 1)
	assertQueueApplicationsRunningMetrics(t, app, 1)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsNewMetrics(t, app, 0)
	// Running -> Completing
	err = app.HandleApplicationEvent(CompleteApplication)
	assertState(t, app, err, Completing.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 0)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 0)
	assertQueueApplicationsCompletingMetrics(t, app, 1)
	// Completing -> Completed
	err = app.HandleApplicationEvent(CompleteApplication)
	assertState(t, app, err, Completed.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 1)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 1)
	assertTotalAppsAcceptedMetrics(t, 1)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 0)
	assertQueueApplicationsCompletedMetrics(t, app, 1)
	assertQueueApplicationsCompletingMetrics(t, app, 0)

	// app-00002: New -> Accepted -> Completing -> Running -> Failing-> Failed
	app = newApplication("app-00002", "default", "root.metrics")
	app.SetQueue(queue)
	assertState(t, app, nil, New.String())
	// New -> Accepted
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Accepted.String())
	// Accepted -> Running
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Running.String())
	// Running -> Completing
	err = app.HandleApplicationEvent(CompleteApplication)
	assertState(t, app, err, Completing.String())
	// Completing -> Running
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Running.String())
	// Running -> Failing
	err = app.HandleApplicationEvent(FailApplication)
	assertState(t, app, err, Failing.String())
	// Failing -> Failed
	err = app.HandleApplicationEvent(FailApplication)
	assertState(t, app, err, Failed.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 1)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 2)
	assertTotalAppsAcceptedMetrics(t, 2)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 1)
	assertQueueApplicationsCompletedMetrics(t, app, 1)

	// app-00003: New -> Accepted -> Running -> Failing -> Failed
	app = newApplication("app-00003", "default", "root.metrics")
	app.SetQueue(queue)
	assertState(t, app, nil, New.String())
	// New -> Accepted
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Accepted.String())
	// Accepted -> Running
	err = app.HandleApplicationEvent(RunApplication)
	assertState(t, app, err, Running.String())
	// Running -> Failing
	err = app.HandleApplicationEvent(FailApplication)
	assertState(t, app, err, Failing.String())
	// Failing -> Failed
	err = app.HandleApplicationEvent(FailApplication)
	assertState(t, app, err, Failed.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 1)
	assertTotalAppsRejectedMetrics(t, 0)
	assertTotalAppsNewMetrics(t, 3)
	assertTotalAppsAcceptedMetrics(t, 3)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 0)
	assertQueueApplicationsFailedMetrics(t, app, 2)
	assertQueueApplicationsCompletedMetrics(t, app, 1)

	// app-00004: New -> Rejected
	app = newApplication("app-00004", "default", "root.metrics")
	app.SetQueue(queue)
	assertState(t, app, nil, New.String())
	// New -> Rejected
	err = app.HandleApplicationEvent(RejectApplication)
	assertState(t, app, err, Rejected.String())
	assertTotalAppsRunningMetrics(t, 0)
	assertTotalAppsCompletedMetrics(t, 1)
	assertTotalAppsRejectedMetrics(t, 1)
	assertTotalAppsNewMetrics(t, 4)
	assertTotalAppsAcceptedMetrics(t, 3)
	assertQueueRunningApps(t, app, 0)
	assertQueueApplicationsNewMetrics(t, app, 0)
	assertQueueApplicationsRunningMetrics(t, app, 0)
	assertQueueApplicationsAcceptedMetrics(t, app, 0)
	assertQueueApplicationsRejectedMetrics(t, app, 1)
	assertQueueApplicationsFailedMetrics(t, app, 2)
	assertQueueApplicationsCompletedMetrics(t, app, 1)

	// app-00005: the queuePath is empty, it will happen for dynamic queue when it before the queue is created
	app = newApplication("app-00005", "default", "")
	assertState(t, app, nil, New.String())
	assertQueueApplicationsNewMetrics(t, app, 0)
	assertTotalAppsNewMetrics(t, 4)
}

func assertState(t testing.TB, app *Application, err error, expected string) {
	t.Helper()
	assert.NilError(t, err, fmt.Sprintf("no error expected when change state to %v", expected))
	assert.Equal(t, app.CurrentState(), expected, "application not in expected state.")
}

func assertTotalAppsNewMetrics(t testing.TB, expected int) {
	t.Helper()
	totalAppsNew, err := metrics.GetSchedulerMetrics().GetTotalApplicationsNew()
	assert.NilError(t, err, "no error expected when getting total new application count.")
	assert.Equal(t, totalAppsNew, expected, "total new application metrics is not as expected.")
}

func assertTotalAppsRunningMetrics(t testing.TB, expected int) {
	t.Helper()
	totalAppsRunning, err := metrics.GetSchedulerMetrics().GetTotalApplicationsRunning()
	assert.NilError(t, err, "no error expected when getting total running application count.")
	assert.Equal(t, totalAppsRunning, expected, "total running application metrics is not as expected.")
}

func assertTotalAppsAcceptedMetrics(t testing.TB, expected int) {
	t.Helper()
	totalAppsAccepted, err := metrics.GetSchedulerMetrics().GetTotalApplicationsAccepted()
	assert.NilError(t, err, "no error expected when getting total accepted application count.")
	assert.Equal(t, totalAppsAccepted, expected, "total accepted application metrics is not as expected.")
}

func assertTotalAppsCompletedMetrics(t testing.TB, expected int) {
	t.Helper()
	totalAppsCompleted, err := metrics.GetSchedulerMetrics().GetTotalApplicationsCompleted()
	assert.NilError(t, err, "no error expected when getting total completed application count.")
	assert.Equal(t, totalAppsCompleted, expected, "total completed application metrics is not as expected.")
}

func assertTotalAppsRejectedMetrics(t testing.TB, expected int) {
	t.Helper()
	totalAppsRejected, err := metrics.GetSchedulerMetrics().GetTotalApplicationsRejected()
	assert.NilError(t, err, "no error expected when getting total rejected application count.")
	assert.Equal(t, totalAppsRejected, expected, "total rejected application metrics is not as expected.")
}

func assertQueueRunningApps(t testing.TB, app *Application, expected int) {
	t.Helper()
	runningApps := app.queue.runningApps
	assert.Equal(t, runningApps, uint64(expected), "total running application in queue is not as expected.") //nolint:gosec
}

func assertQueueApplicationsAcceptedMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsAccepted, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsAccepted()
	assert.NilError(t, err, "no error expected when getting total accepted application count in queue.")
	assert.Equal(t, queueApplicationsAccepted, expected, "total accepted application metrics in queue is not as expected.")
}

func assertQueueApplicationsRejectedMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsRejected, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsRejected()
	assert.NilError(t, err, "no error expected when getting total rejected application count in queue.")
	assert.Equal(t, queueApplicationsRejected, expected, "total rejected application metrics in queue is not as expected.")
}

func assertQueueApplicationsRunningMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsRunning, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsRunning()
	assert.NilError(t, err, "no error expected when getting total running application count in queue.")
	assert.Equal(t, queueApplicationsRunning, expected, "total running application metrics in queue is not as expected.")
}

func assertQueueApplicationsFailedMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsFailed, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsFailed()
	assert.NilError(t, err, "no error expected when getting total failed application count in queue.")
	assert.Equal(t, queueApplicationsFailed, expected, "total failed application metrics in queue is not as expected.")
}

func assertQueueApplicationsResumingMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsResuming, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsResuming()
	assert.NilError(t, err, "no error expected when getting total resuming application count in queue.")
	assert.Equal(t, queueApplicationsResuming, expected, "total resuming application metrics in queue is not as expected.")
}

func assertQueueApplicationsCompletedMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsCompleted, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsCompleted()
	assert.NilError(t, err, "no error expected when getting total completed application count in queue.")
	assert.Equal(t, queueApplicationsCompleted, expected, "total completed application metrics in queue is not as expected.")
}

func assertQueueApplicationsCompletingMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsCompleting, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsCompleting()
	assert.NilError(t, err, "no error expected when getting total completing application count in queue.")
	assert.Equal(t, queueApplicationsCompleting, expected, "total completing application metrics in queue is not as expected.")
}

func assertQueueApplicationsNewMetrics(t testing.TB, app *Application, expected int) {
	t.Helper()
	queueApplicationsNew, err := metrics.GetQueueMetrics(app.queuePath).GetQueueApplicationsNew()
	assert.NilError(t, err, "no error expected when getting total new applications in queue.")
	assert.Equal(t, queueApplicationsNew, expected, "total new applications in queue is not as expected.")
}

func createQueue(t *testing.T, queueName string) *Queue {
	root, err := createRootQueue(nil)
	assert.NilError(t, err, "failed to create queue: %v", err)
	queue, err := createManagedQueue(root, queueName, false, map[string]string{"cpu": "10"})
	assert.NilError(t, err, "failed to create queue: %v", err)
	return queue
}

func isStateChangeEvent(t *testing.T, app *Application, changeDetail si.EventRecord_ChangeDetail, record *si.EventRecord) {
	assert.Equal(t, si.EventRecord_APP, record.Type, "incorrect event type, expect app")
	assert.Equal(t, app.ApplicationID, record.ObjectID, "incorrect object ID, expected application ID")
	assert.Equal(t, si.EventRecord_SET, record.EventChangeType, "incorrect change type, expected set")
	assert.Equal(t, changeDetail, record.EventChangeDetail, "incorrect change detail")
}

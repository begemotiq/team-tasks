//go:build integration

package integration

import (
	"testing"
)

type reportDataset struct {
	owner       int64
	admin       int64
	member      int64
	team        int64
	doneTask    int64
	invalidTask int64
}

func TestMySQLRepositoryReports(t *testing.T) {
	t.Parallel()

	fixture := newFixture(t)
	dataset := fixture.reportDataset()

	summary, err := fixture.repos.tasks.TeamSummary(fixture.ctx, dataset.owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 || summary[0].TeamID != dataset.team || summary[0].MembersCount != 3 || summary[0].DoneTasksLast7Days != 1 {
		t.Fatalf("unexpected owner summary: %#v", summary)
	}

	summary, err = fixture.repos.tasks.TeamSummary(fixture.ctx, dataset.admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 || summary[0].TeamID != dataset.team {
		t.Fatalf("unexpected admin summary: %#v", summary)
	}

	summary, err = fixture.repos.tasks.TeamSummary(fixture.ctx, dataset.member)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 0 {
		t.Fatalf("member must not receive management summary: %#v", summary)
	}

	top, err := fixture.repos.tasks.TopCreatorsByTeam(fixture.ctx, dataset.owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) < 2 {
		t.Fatalf("unexpected top creators: %#v", top)
	}
	if top[0].TeamID != dataset.team || top[0].UserID != dataset.owner || top[0].TasksCreated != 2 || top[0].RankPosition != 1 {
		t.Fatalf("unexpected first top creator: %#v", top[0])
	}
	if top[1].TeamID != dataset.team || top[1].UserID != dataset.member || top[1].TasksCreated != 1 || top[1].RankPosition != 2 {
		t.Fatalf("unexpected second top creator: %#v", top[1])
	}

	invalid, err := fixture.repos.tasks.InvalidAssignees(fixture.ctx, dataset.admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(invalid) != 1 || invalid[0].ID != dataset.invalidTask {
		t.Fatalf("unexpected invalid assignees: %#v", invalid)
	}
	if invalid[0].ID == dataset.doneTask {
		t.Fatal("valid assigned task was returned as invalid")
	}
}

func (f *mysqlFixture) reportDataset() reportDataset {
	f.t.Helper()

	f.loadSQLFixture("report_dataset.sql", map[string]string{
		"{{suffix}}": f.suffix,
	})

	owner := f.mustUserByEmail("owner-" + f.suffix + "@example.com")
	admin := f.mustUserByEmail("admin-" + f.suffix + "@example.com")
	member := f.mustUserByEmail("member-" + f.suffix + "@example.com")
	team := f.mustTeamByName(owner, "backend "+f.suffix)
	doneTask := f.mustTaskByTitle(owner, team, "done "+f.suffix)
	invalidTask := f.mustTaskByTitle(owner, team, "invalid "+f.suffix)

	return reportDataset{
		owner:       owner.ID,
		admin:       admin.ID,
		member:      member.ID,
		team:        team.ID,
		doneTask:    doneTask.ID,
		invalidTask: invalidTask.ID,
	}
}

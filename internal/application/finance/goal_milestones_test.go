package finance

import "testing"

func TestMilestoneCrossThenDrop(t *testing.T) {
	crossed := planGoalMilestones(nil, 50)
	if len(crossed.Insert) != 2 || crossed.Insert[0] != 25 || crossed.Insert[1] != 50 {
		t.Fatalf("insert %+v", crossed.Insert)
	}
	if len(crossed.Uncelebrated) != 2 {
		t.Fatalf("uncelebrated %+v", crossed.Uncelebrated)
	}

	existing := []milestoneRow{
		{Percent: 25, Celebrated: true},
		{Percent: 50, Celebrated: false},
	}
	dropped := planGoalMilestones(existing, 20)
	if len(dropped.Delete) != 2 || len(dropped.Uncelebrated) != 0 || len(dropped.Insert) != 0 {
		t.Fatalf("drop %+v", dropped)
	}

	again := planGoalMilestones(nil, 25)
	if len(again.Insert) != 1 || again.Insert[0] != 25 || len(again.Uncelebrated) != 1 {
		t.Fatalf("recross %+v", again)
	}
}

func TestMilestoneKeepsCelebratedThreshold(t *testing.T) {
	plan := planGoalMilestones([]milestoneRow{{Percent: 25, Celebrated: true}}, 30)
	if len(plan.Insert) != 0 || len(plan.Delete) != 0 || len(plan.Uncelebrated) != 0 {
		t.Fatalf("celebrated row should stay quiet: %+v", plan)
	}
}

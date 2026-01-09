package festival

import "testing"

func TestBuildElementName_StripsNumericPrefix(t *testing.T) {
	tests := []struct {
		number   int
		name     string
		elemType ElementType
		want     string
	}{
		{1, "001_PLANNING", PhaseType, "001_PLANNING"},
		{2, "002 planning", PhaseType, "002_PLANNING"},
		{3, "003-Review", PhaseType, "003_REVIEW"},
		{1, "01_requirements", SequenceType, "01_requirements"},
		{2, "02 requirements", SequenceType, "02_requirements"},
		{3, "03-Backend API", SequenceType, "03_backend_api"},
		{4, "04_Task_Name", TaskType, "04_task_name"},
		{5, "123abc", PhaseType, "005_123ABC"},
		{7, "QA & Review", PhaseType, "007_QA_&_REVIEW"},
		{12, "Very Long Phase Name For Naming Coverage", PhaseType, "012_VERY_LONG_PHASE_NAME_FOR_NAMING_COVERAGE"},
		{99, "API v2", SequenceType, "99_api_v2"},
		{7, "User-Flow Edge", SequenceType, "07_user-flow_edge"},
	}

	for _, tt := range tests {
		got := BuildElementName(tt.number, tt.name, tt.elemType)
		if got != tt.want {
			t.Errorf("BuildElementName(%d, %q, %s) = %q, want %q", tt.number, tt.name, tt.elemType.String(), got, tt.want)
		}
	}
}

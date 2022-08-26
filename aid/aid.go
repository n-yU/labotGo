// アクションID管理
package aid

const (
	// メンバー管理
	BaseMember             = "Member"
	AddMember              = "MemberAddMember"
	AddMemberSelectUser    = "MemberAddMemberSelectUser"
	AddMemberSelectTeams   = "MemberAddMemberSelectTeams"
	EditMember             = "MemberEditMember"
	EditMemberSelectMember = "MemberEditMemberSelectMember"
	EditMemberSelectTeams  = "MemberEditMemberSelectTeams"

	// チーム管理
	BaseTeam              = "Team"
	AddTeam               = "TeamAddTeam"
	AddTeamInputName      = "TeamAddTeamInputName"
	AddTeamSelectMembers  = "TeamAddTeamSelectMembers"
	EditTeam              = "TeamEditTeam"
	EditTeamInputName     = "TeamEditTeamInputName"
	EditTeamSelectMembers = "TeamEditTeamSelectMembers"
)

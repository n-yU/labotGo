// アクションID管理
package aid

const (
	// メンバー管理
	BaseMember               = "Member"
	AddMember                = "MemberAddMember"
	AddMemberSelectUser      = "MemberAddMemberSelectUser"
	AddMemberSelectTeams     = "MemberAddMemberSelectTeams"
	EditMember               = "MemberEditMember"
	EditMemberSelectMember   = "MemberEditMemberSelectMember"
	EditMemberSelectTeams    = "MemberEditMemberSelectTeams"
	DeleteMember             = "MemberDeleteMember"
	DeleteMemberSelectMember = "MemberDeleteMemberSelectMember"

	// チーム管理
	BaseTeam              = "Team"
	AddTeam               = "TeamAddTeam"
	AddTeamInputName      = "TeamAddTeamInputName"
	AddTeamSelectMembers  = "TeamAddTeamSelectMembers"
	EditTeam              = "TeamEditTeam"
	EditTeamSelectName    = "TeamEditTeamSelectName"
	EditTeamInputName     = "TeamEditTeamInputName"
	EditTeamSelectMembers = "TeamEditTeamSelectMembers"
	DeleteTeam            = "TeamDeleteTeam"
	DeleteTeamSelectTeam  = "TeamDeleteTeamSelectTeam"

	// グルーピング
	BaseGroup                = "Group"
	GroupTeam                = "GroupTeam"
	GroupTeamSelectNames     = "GroupTeamSelectNames"
	GroupTeamSelectType      = "GroupTeamSelectType"
	GroupTeamInputValue      = "GroupTeamInputValue"
	GroupCustom              = "GroupCustom"
	GroupCustomSelectMembers = "GroupCustomSelectMembers"
	GroupCustomSelectType    = "GroupCustomSelectType"
	GroupCustomInputValue    = "GroupCustomInputValue"
)

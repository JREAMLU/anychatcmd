package ui

import (
	"github.com/hawklithm/anychatcmd/utils"
	"github.com/hawklithm/anychatcmd/wechat"
	ui "github.com/hawklithm/termui"
	"github.com/hawklithm/termui/widgets"
	"log"
	"strings"
	"time"
)

type UserList struct {
	currentId       string
	curUserIndex    int
	logger          *log.Logger
	baseX           int
	baseY           int
	width           int
	height          int
	recentUserList  []UserInfo
	recentGroupList []*Group
	userList        *SortItems
	groupList       *SortItems
	selectEvent     chan SelectEvent
	we              *wechat.Wechat

	tabPane          *widgets.TabPane
	userNickListBox  *widgets.List
	groupNickListBox *widgets.List
	recentListBox    *widgets.List
	tabWidgets       []*widgets.List
	tabLists         []*SortItems
	picked           bool
	currentTab       *widgets.List
	recentList       *SortItems
	currentList      *SortItems
}

func (this *UserList) Pick() {
	this.picked = true
	this.currentTab.BorderStyle = ui.NewStyle(ui.ColorRed)
	ui.Render(this.currentTab)
}

func (this *UserList) Unpick() {
	this.picked = false
	this.currentTab.BorderStyle = ui.NewStyle(ui.ColorMagenta)
	ui.Render(this.currentTab)
}

type UserChangeEvent struct {
}

type SelectEvent interface {
	GetId() string
	GetName() string
	GetType() string
	GetLastChatTime() time.Time
	GetUserList() []*UserInfo
}

type UserInfo struct {
	UserId       string
	Nick         string
	DisplayName  string
	LastChatTime time.Time
}

func (l UserInfo) GetUserList() []*UserInfo {
	return []*UserInfo{&l}
}

func (l UserInfo) GetLastChatTime() time.Time {
	return l.LastChatTime
}

func (l UserInfo) GetId() string {
	return l.UserId
}

func (l UserInfo) GetName() string {
	return utils.If(l.DisplayName != "", l.DisplayName, l.Nick).(string)
}

func (l UserInfo) GetType() string {
	return "user"
}

//用户群组
type Group struct {
	UserList     []*UserInfo
	Name         string
	GroupId      string
	LastChatTime time.Time
}

func (l Group) GetUserList() []*UserInfo {
	return l.UserList
}

func (l Group) GetLastChatTime() time.Time {
	return l.LastChatTime
}

func (l Group) GetId() string {
	return l.GroupId
}

func (l Group) GetName() string {
	return l.Name
}

func (l Group) GetType() string {
	return "group"
}

type SortItems []SelectEvent

func (p SortItems) Len() int { return len(p) }

func (p SortItems) Less(i, j int) bool {
	return p[i].GetLastChatTime().After(p[j].GetLastChatTime())
}

func merge(userInfo []UserInfo, groupInfo []*Group) *SortItems {
	z := make(SortItems, len(userInfo)+len(groupInfo))
	var count = 0
	for _, user := range userInfo {
		z[count] = user
		count++
	}
	for _, group := range groupInfo {
		z[count] = group
		count++
	}
	return &z
}

func (this *UserList) renderTab() {
	this.currentTab = this.tabWidgets[this.tabPane.ActiveTabIndex]
	this.currentList = this.tabLists[this.tabPane.ActiveTabIndex]
	this.refreshCurrentSelect()
	ui.Render(this.tabPane, this.currentTab)
}

func InitTalkInfo(w *wechat.Wechat, logger *log.Logger, groups []*Group) {
	var groupIds []string
	groupMap := make(map[string]*Group)
	for _, group := range groups {
		groupIds = append(groupIds, group.GroupId)
		groupMap[group.GroupId] = group
	}
	l := len(groupIds)
	var end int
	for index := 0; index < l; index += 50 {
		end = index + 50
		if end > l {
			end = l
		}
		if members, err := w.GetContactsInBatch(groupIds[index:end]); err != nil {
			logger.Println("init contact error!", err)
		} else {
			if members != nil && len(members) > 0 {
				for _, member := range members {
					groupMap[member.UserName].
						UserList = convertMemeberToUserInfo(member.MemberList)
				}

			}
		}
	}
}

func convertMemeberToUserInfo(users []wechat.User) []*UserInfo {
	var userInfos []*UserInfo
	if users == nil || len(users) == 0 {
		return userInfos
	}
	for _, u := range users {
		userInfos = append(userInfos, &UserInfo{
			UserId:      u.UserName,
			Nick:        u.NickName,
			DisplayName: u.RemarkName})
	}
	return userInfos
}

func (this *UserList) refreshCurrentSelect() {
	this.curUserIndex = this.currentTab.SelectedRow
	if this.selectEvent != nil {
		u := (*this.currentList)[this.curUserIndex]
		if strings.HasPrefix(u.GetId(), "@@") && len(u.GetUserList()) == 0 {
			InitTalkInfo(this.we, this.logger, []*Group{u.(*Group)})
		}
		this.selectEvent <- u
	} else {
		this.logger.Println("warning!", "no select event channel set!")
	}
}

func (this *UserList) Reset() {
	this.userNickListBox = widgets.NewList()
	//userNickListBox.Title = "用户列表"
	//userNickListBox.BorderStyle = ui.NewStyle(ui.ColorMagenta)
	//userNickListBox.Border = true
	this.userNickListBox.TextStyle = ui.NewStyle(ui.ColorYellow)
	this.userNickListBox.WrapText = false
	this.userNickListBox.SelectedRowStyle = ui.NewStyle(ui.ColorWhite,
		ui.ColorRed)

	this.userNickListBox.SetRect(this.baseX, this.baseY+2,
		this.baseX+this.width,
		this.baseY+this.height)

	groupNickListBox := widgets.NewList()
	//userNickListBox.Title = "用户列表"
	//userNickListBox.BorderStyle = ui.NewStyle(ui.ColorMagenta)
	//userNickListBox.Border = true
	groupNickListBox.TextStyle = ui.NewStyle(ui.ColorYellow)
	groupNickListBox.WrapText = false
	groupNickListBox.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorRed)

	groupNickListBox.SetRect(this.baseX, this.baseY+2, this.baseX+this.width,
		this.baseY+this.height)

	this.groupNickListBox = groupNickListBox

	this.recentListBox = widgets.NewList()
	this.recentListBox.TextStyle = ui.NewStyle(ui.ColorYellow)
	this.recentListBox.WrapText = false
	this.recentListBox.SelectedRowStyle = ui.NewStyle(ui.ColorWhite,
		ui.ColorRed)

	this.recentListBox.SetRect(this.baseX, this.baseY+2,
		this.baseX+this.width,
		this.baseY+this.height)

	nickList := make([]string, len(*this.userList))
	groupList := make([]string, len(*this.groupList))

	for i, user := range *this.userList {
		nickList[i] = user.GetName()
	}
	this.userNickListBox.Rows = nickList

	for i, group := range *this.groupList {
		groupList[i] = group.GetName()
	}
	this.groupNickListBox.Rows = groupList

	this.recentList = merge(this.recentUserList, this.recentGroupList)
	recentNickList := make([]string, this.recentList.Len())
	for i, r := range *this.recentList {
		recentNickList[i] = r.GetName()
	}
	this.recentListBox.Rows = recentNickList

	this.tabPane = widgets.NewTabPane("聊天列表", "好友列表", "群列表")
	this.tabPane.SetRect(this.baseX, this.baseY, this.baseX+this.width,
		this.baseY+3)
	this.tabPane.Border = true
	tabWidgets := make([]*widgets.List, 3)
	tabWidgets[0] = this.recentListBox
	tabWidgets[1] = this.userNickListBox
	tabWidgets[2] = this.groupNickListBox
	this.tabWidgets = tabWidgets

	this.tabLists = make([]*SortItems, 3)
	this.tabLists[0] = this.recentList
	this.tabLists[1] = this.userList
	this.tabLists[2] = this.groupList

	this.currentTab = this.recentListBox
	this.currentList = this.recentList

	ui.Render(this.tabPane, this.currentTab)
}

func convertUsersToSortItems(users []UserInfo) *SortItems {
	z := make(SortItems, len(users))
	for i, user := range users {
		z[i] = user
	}
	return &z
}

func convertGroupsToSortItems(groups []*Group) *SortItems {
	z := make(SortItems, len(groups))
	for i, user := range groups {
		z[i] = user
	}
	return &z
}

func NewUserList(recentUserList []UserInfo, recentGroupList []*Group,
	userList []UserInfo,
	groupList []*Group,
	selectEvent chan SelectEvent, width,
	height,
	baseX, baseY int,
	logger *log.Logger, we *wechat.Wechat) *UserList {

	//	chinese := false
	convertUsersToSortItems(userList)

	l := &UserList{
		curUserIndex:    0,
		logger:          logger,
		baseX:           baseX,
		baseY:           baseY,
		width:           width,
		height:          height,
		recentUserList:  recentUserList,
		recentGroupList: recentGroupList,
		userList:        convertUsersToSortItems(userList),
		groupList:       convertGroupsToSortItems(groupList),
		selectEvent:     selectEvent,
		we:              we,
	}

	l.Reset()

	return l

}

func (l *UserList) InvokeSelect() {
	l.refreshCurrentSelect()
}

func (l *UserList) nextUser() {
	l.currentTab.ScrollDown()
	l.refreshCurrentSelect()
	ui.Render(l.currentTab)
}

func (l *UserList) prevUser() {
	l.currentTab.ScrollUp()
	l.refreshCurrentSelect()
	ui.Render(l.currentTab)
}

func (l *UserList) focuseLeft() {
	l.tabPane.FocusLeft()
	l.renderTab()
}

func (l *UserList) focuseRight() {
	l.tabPane.FocusRight()
	l.renderTab()
}

func (this *UserList) Action(e ui.Event) bool {
	if !this.picked {
		return false
	}
	switch e.ID {
	case "j":
		this.nextUser()
		return true
	case "k":
		this.prevUser()
		return true
	case "h":
		this.focuseLeft()
		return true
	case "l":
		this.focuseRight()
		return true
	}
	return false

}

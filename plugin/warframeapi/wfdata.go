package warframeapi

import "time"

type WFAPI struct {
	Timestamp            time.Time            `json:"timestamp"`
	News                 []News               `json:"news"`
	Events               []Events             `json:"events"`
	Alerts               []Alerts             `json:"alerts"`
	Sortie               Sortie               `json:"sortie"`
	SyndicateMissions    []SyndicateMissions  `json:"syndicateMissions"`
	Fissures             []Fissures           `json:"fissures"`
	GlobalUpgrades       []interface{}        `json:"globalUpgrades"`
	FlashSales           []FlashSales         `json:"flashSales"`
	Invasions            []Invasions          `json:"invasions"`
	DarkSectors          []interface{}        `json:"darkSectors"`
	VoidTrader           VoidTrader           `json:"voidTrader"`
	DailyDeals           []DailyDeals         `json:"dailyDeals"`
	Simaris              Simaris              `json:"simaris"`
	ConclaveChallenges   []ConclaveChallenges `json:"conclaveChallenges"`
	PersistentEnemies    []interface{}        `json:"persistentEnemies"`
	EarthCycle           EarthCycle           `json:"earthCycle"`
	CetusCycle           CetusCycle           `json:"cetusCycle"`
	CambionCycle         CambionCycle         `json:"cambionCycle"`
	ZarimanCycle         ZarimanCycle         `json:"zarimanCycle"`
	WeeklyChallenges     []interface{}        `json:"weeklyChallenges"`
	ConstructionProgress ConstructionProgress `json:"constructionProgress"`
	VallisCycle          VallisCycle          `json:"vallisCycle"`
	Nightwave            Nightwave            `json:"nightwave"`
	Kuva                 []interface{}        `json:"kuva"`
	Arbitration          Arbitration          `json:"arbitration"`
	SentientOutposts     SentientOutposts     `json:"sentientOutposts"`
	SteelPath            SteelPath            `json:"steelPath"`
	VaultTrader          VaultTrader          `json:"vaultTrader"`
}
type Translations struct {
	En string `json:"en"`
	Fr string `json:"fr"`
	It string `json:"it"`
	De string `json:"de"`
	Es string `json:"es"`
	Pt string `json:"pt"`
	Ru string `json:"ru"`
	Pl string `json:"pl"`
	Uk string `json:"uk"`
	Tr string `json:"tr"`
	Ja string `json:"ja"`
	Zh string `json:"zh"`
	Ko string `json:"ko"`
	Tc string `json:"tc"`
}
type News struct {
	ID           string       `json:"id"`
	Message      string       `json:"message"`
	Link         string       `json:"link"`
	ImageLink    string       `json:"imageLink"`
	Priority     bool         `json:"priority"`
	Date         time.Time    `json:"date"`
	Eta          string       `json:"eta"`
	Update       bool         `json:"update"`
	PrimeAccess  bool         `json:"primeAccess"`
	Stream       bool         `json:"stream"`
	Translations Translations `json:"translations"`
	AsString     string       `json:"asString"`
}
type Metadata struct {
}
type NextAlt struct {
	Expiry     time.Time `json:"expiry"`
	Activation time.Time `json:"activation"`
}
type Events struct {
	ID                string        `json:"id"`
	Activation        time.Time     `json:"activation"`
	StartString       string        `json:"startString"`
	Expiry            time.Time     `json:"expiry"`
	Active            bool          `json:"active"`
	MaximumScore      int           `json:"maximumScore"`
	CurrentScore      int           `json:"currentScore"`
	SmallInterval     interface{}   `json:"smallInterval"`
	LargeInterval     interface{}   `json:"largeInterval"`
	Faction           string        `json:"faction"`
	Description       string        `json:"description"`
	Tooltip           string        `json:"tooltip"`
	Node              string        `json:"node"`
	ConcurrentNodes   []interface{} `json:"concurrentNodes"`
	Rewards           []interface{} `json:"rewards"`
	Expired           bool          `json:"expired"`
	InterimSteps      []interface{} `json:"interimSteps"`
	ProgressSteps     []interface{} `json:"progressSteps"`
	IsPersonal        bool          `json:"isPersonal"`
	RegionDrops       []interface{} `json:"regionDrops"`
	ArchwingDrops     []interface{} `json:"archwingDrops"`
	AsString          string        `json:"asString"`
	Metadata          Metadata      `json:"metadata"`
	CompletionBonuses []interface{} `json:"completionBonuses"`
	AltExpiry         time.Time     `json:"altExpiry"`
	AltActivation     time.Time     `json:"altActivation"`
	NextAlt           NextAlt       `json:"nextAlt"`
}
type Variants struct {
	MissionType         string `json:"missionType"`
	Modifier            string `json:"modifier"`
	ModifierDescription string `json:"modifierDescription"`
	Node                string `json:"node"`
}
type Sortie struct {
	ID          string     `json:"id"`
	Activation  time.Time  `json:"activation"`
	StartString string     `json:"startString"`
	Expiry      time.Time  `json:"expiry"`
	Active      bool       `json:"active"`
	RewardPool  string     `json:"rewardPool"`
	Variants    []Variants `json:"variants"`
	Boss        string     `json:"boss"`
	Faction     string     `json:"faction"`
	Expired     bool       `json:"expired"`
	Eta         string     `json:"eta"`
}
type Jobs struct {
	ID             string    `json:"id"`
	RewardPool     []string  `json:"rewardPool"`
	Type           string    `json:"type"`
	EnemyLevels    []int     `json:"enemyLevels"`
	StandingStages []int     `json:"standingStages"`
	MinMR          int       `json:"minMR"`
	Expiry         time.Time `json:"expiry"`
	TimeBound      string    `json:"timeBound,omitempty"`
}
type SyndicateMissions struct {
	ID           string        `json:"id"`
	Activation   time.Time     `json:"activation"`
	StartString  string        `json:"startString"`
	Expiry       time.Time     `json:"expiry"`
	Active       bool          `json:"active"`
	Syndicate    string        `json:"syndicate"`
	SyndicateKey string        `json:"syndicateKey"`
	Nodes        []interface{} `json:"nodes"`
	Jobs         []Jobs        `json:"jobs"`
	Eta          string        `json:"eta"`
}
type Fissures struct {
	ID          string    `json:"id"`
	Activation  time.Time `json:"activation"`
	StartString string    `json:"startString"`
	Expiry      time.Time `json:"expiry"`
	Active      bool      `json:"active"`
	Node        string    `json:"node"`
	MissionType string    `json:"missionType"`
	MissionKey  string    `json:"missionKey"`
	Enemy       string    `json:"enemy"`
	EnemyKey    string    `json:"enemyKey"`
	NodeKey     string    `json:"nodeKey"`
	Tier        string    `json:"tier"`
	TierNum     int       `json:"tierNum"`
	Expired     bool      `json:"expired"`
	Eta         string    `json:"eta"`
	IsStorm     bool      `json:"isStorm"`
}
type FlashSales struct {
	Item            string    `json:"item"`
	Expiry          time.Time `json:"expiry"`
	Activation      time.Time `json:"activation"`
	Discount        int       `json:"discount"`
	RegularOverride int       `json:"regularOverride"`
	PremiumOverride int       `json:"premiumOverride"`
	IsShownInMarket bool      `json:"isShownInMarket"`
	IsFeatured      bool      `json:"isFeatured"`
	IsPopular       bool      `json:"isPopular"`
	ID              string    `json:"id"`
	Expired         bool      `json:"expired"`
	Eta             string    `json:"eta"`
}
type CountedItems struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
	Key   string `json:"key"`
}
type AttackerReward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []CountedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type Reward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []CountedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type Attacker struct {
	Reward     Reward `json:"reward"`
	Faction    string `json:"faction"`
	FactionKey string `json:"factionKey"`
}
type DefenderReward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []CountedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type Defender struct {
	Reward     Reward `json:"reward"`
	Faction    string `json:"faction"`
	FactionKey string `json:"factionKey"`
}
type Invasions struct {
	ID               string         `json:"id"`
	Activation       time.Time      `json:"activation"`
	StartString      string         `json:"startString"`
	Node             string         `json:"node"`
	NodeKey          string         `json:"nodeKey"`
	Desc             string         `json:"desc"`
	AttackerReward   AttackerReward `json:"attackerReward"`
	AttackingFaction string         `json:"attackingFaction"`
	Attacker         Attacker       `json:"attacker"`
	DefenderReward   DefenderReward `json:"defenderReward"`
	DefendingFaction string         `json:"defendingFaction"`
	Defender         Defender       `json:"defender"`
	VsInfestation    bool           `json:"vsInfestation"`
	Count            int            `json:"count"`
	RequiredRuns     int            `json:"requiredRuns"`
	Completion       float64        `json:"completion"`
	Completed        bool           `json:"completed"`
	Eta              string         `json:"eta"`
	RewardTypes      []string       `json:"rewardTypes"`
}
type VoidTrader struct {
	ID           string        `json:"id"`
	Activation   time.Time     `json:"activation"`
	StartString  string        `json:"startString"`
	Expiry       time.Time     `json:"expiry"`
	Active       bool          `json:"active"`
	Character    string        `json:"character"`
	Location     string        `json:"location"`
	Inventory    []interface{} `json:"inventory"`
	PsID         string        `json:"psId"`
	EndString    string        `json:"endString"`
	InitialStart time.Time     `json:"initialStart"`
	Schedule     []interface{} `json:"schedule"`
}
type DailyDeals struct {
	Item          string    `json:"item"`
	Expiry        time.Time `json:"expiry"`
	Activation    time.Time `json:"activation"`
	OriginalPrice int       `json:"originalPrice"`
	SalePrice     int       `json:"salePrice"`
	Total         int       `json:"total"`
	Sold          int       `json:"sold"`
	ID            string    `json:"id"`
	Eta           string    `json:"eta"`
	Discount      int       `json:"discount"`
}
type Simaris struct {
	Target         string `json:"target"`
	IsTargetActive bool   `json:"isTargetActive"`
	AsString       string `json:"asString"`
}
type ConclaveChallenges struct {
	ID            string    `json:"id"`
	Expiry        time.Time `json:"expiry"`
	Activation    time.Time `json:"activation"`
	Amount        int       `json:"amount"`
	Mode          string    `json:"mode"`
	Category      string    `json:"category"`
	Eta           string    `json:"eta"`
	Expired       bool      `json:"expired"`
	Daily         bool      `json:"daily"`
	RootChallenge bool      `json:"rootChallenge"`
	EndString     string    `json:"endString"`
	Description   string    `json:"description"`
	Title         string    `json:"title"`
	Standing      int       `json:"standing"`
	AsString      string    `json:"asString"`
}
type EarthCycle struct {
	ID         string    `json:"id"`
	Expiry     time.Time `json:"expiry"`
	Activation time.Time `json:"activation"`
	IsDay      bool      `json:"isDay"`
	State      string    `json:"state"`
	TimeLeft   string    `json:"timeLeft"`
}
type CetusCycle struct {
	ID          string    `json:"id"`
	Expiry      time.Time `json:"expiry"`
	Activation  time.Time `json:"activation"`
	IsDay       bool      `json:"isDay"`
	State       string    `json:"state"`
	TimeLeft    string    `json:"timeLeft"`
	IsCetus     bool      `json:"isCetus"`
	ShortString string    `json:"shortString"`
}
type CambionCycle struct {
	ID         string    `json:"id"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
	TimeLeft   string    `json:"timeLeft"`
	Active     string    `json:"active"`
}
type ZarimanCycle struct {
	ID              string    `json:"id"`
	BountiesEndDate time.Time `json:"bountiesEndDate"`
	Expiry          time.Time `json:"expiry"`
	Activation      time.Time `json:"activation"`
	IsCorpus        bool      `json:"isCorpus"`
	State           string    `json:"state"`
	TimeLeft        string    `json:"timeLeft"`
	ShortString     string    `json:"shortString"`
}
type ConstructionProgress struct {
	ID                string `json:"id"`
	FomorianProgress  string `json:"fomorianProgress"`
	RazorbackProgress string `json:"razorbackProgress"`
	UnknownProgress   string `json:"unknownProgress"`
}
type VallisCycle struct {
	ID          string    `json:"id"`
	Expiry      time.Time `json:"expiry"`
	IsWarm      bool      `json:"isWarm"`
	State       string    `json:"state"`
	Activation  time.Time `json:"activation"`
	TimeLeft    string    `json:"timeLeft"`
	ShortString string    `json:"shortString"`
}
type Params struct {
}
type ActiveChallenges struct {
	ID          string    `json:"id"`
	Activation  time.Time `json:"activation"`
	StartString string    `json:"startString"`
	Expiry      time.Time `json:"expiry"`
	Active      bool      `json:"active"`
	IsDaily     bool      `json:"isDaily,omitempty"`
	IsElite     bool      `json:"isElite"`
	Desc        string    `json:"desc"`
	Title       string    `json:"title"`
	Reputation  int       `json:"reputation"`
}
type Nightwave struct {
	ID                 string             `json:"id"`
	Activation         time.Time          `json:"activation"`
	StartString        string             `json:"startString"`
	Expiry             time.Time          `json:"expiry"`
	Active             bool               `json:"active"`
	Season             int                `json:"season"`
	Tag                string             `json:"tag"`
	Phase              int                `json:"phase"`
	Params             Params             `json:"params"`
	PossibleChallenges []interface{}      `json:"possibleChallenges"`
	ActiveChallenges   []ActiveChallenges `json:"activeChallenges"`
	RewardTypes        []string           `json:"rewardTypes"`
}
type Arbitration struct {
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
	Enemy      string    `json:"enemy"`
	Type       string    `json:"type"`
	Archwing   bool      `json:"archwing"`
	Sharkwing  bool      `json:"sharkwing"`
	Node       string    `json:"node"`
	NodeKey    string    `json:"nodeKey"`
	TypeKey    string    `json:"typeKey"`
	ID         string    `json:"id"`
	Expired    bool      `json:"expired"`
}
type Mission struct {
	Node    string `json:"node"`
	Faction string `json:"faction"`
	Type    string `json:"type"`
}
type SentientOutposts struct {
	Mission    Mission   `json:"mission"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
	Active     bool      `json:"active"`
	ID         string    `json:"id"`
}
type CurrentReward struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type Rotation struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type Evergreens struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type Incursions struct {
	ID         string    `json:"id"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
}
type SteelPath struct {
	CurrentReward CurrentReward `json:"currentReward"`
	Activation    time.Time     `json:"activation"`
	Expiry        time.Time     `json:"expiry"`
	Remaining     string        `json:"remaining"`
	Rotation      []Rotation    `json:"rotation"`
	Evergreens    []Evergreens  `json:"evergreens"`
	Incursions    Incursions    `json:"incursions"`
}
type Inventory struct {
	Item    string      `json:"item"`
	Ducats  int         `json:"ducats"`
	Credits interface{} `json:"credits"`
}
type Schedule struct {
	Expiry time.Time `json:"expiry"`
	Item   string    `json:"item"`
}
type VaultTrader struct {
	ID           string      `json:"id"`
	Activation   time.Time   `json:"activation"`
	StartString  string      `json:"startString"`
	Expiry       time.Time   `json:"expiry"`
	Active       bool        `json:"active"`
	Character    string      `json:"character"`
	Location     string      `json:"location"`
	Inventory    []Inventory `json:"inventory"`
	PsID         string      `json:"psId"`
	EndString    string      `json:"endString"`
	InitialStart time.Time   `json:"initialStart"`
	Completed    bool        `json:"completed"`
	Schedule     []Schedule  `json:"schedule"`
}
type Alerts struct {
	ID          string    `json:"id"`
	Activation  time.Time `json:"activation"`
	StartString string    `json:"startString"`
	Expiry      time.Time `json:"expiry"`
	Active      bool      `json:"active"`
	Mission     struct {
		Description string `json:"description"`
		Node        string `json:"node"`
		NodeKey     string `json:"nodeKey"`
		Type        string `json:"type"`
		TypeKey     string `json:"typeKey"`
		Faction     string `json:"faction"`
		Reward      struct {
			Items        []interface{} `json:"items"`
			CountedItems []struct {
				Count int    `json:"count"`
				Type  string `json:"type"`
				Key   string `json:"key"`
			} `json:"countedItems"`
			Credits    int    `json:"credits"`
			AsString   string `json:"asString"`
			ItemString string `json:"itemString"`
			Thumbnail  string `json:"thumbnail"`
			Color      int    `json:"color"`
		} `json:"reward"`
		MinEnemyLevel    int           `json:"minEnemyLevel"`
		MaxEnemyLevel    int           `json:"maxEnemyLevel"`
		MaxWaveNum       int           `json:"maxWaveNum"`
		Nightmare        bool          `json:"nightmare"`
		ArchwingRequired bool          `json:"archwingRequired"`
		IsSharkwing      bool          `json:"isSharkwing"`
		LevelOverride    string        `json:"levelOverride"`
		EnemySpec        string        `json:"enemySpec"`
		AdvancedSpawners []interface{} `json:"advancedSpawners"`
		RequiredItems    []interface{} `json:"requiredItems"`
		LevelAuras       []interface{} `json:"levelAuras"`
	} `json:"mission"`
	Eta         string   `json:"eta"`
	RewardTypes []string `json:"rewardTypes"`
	Tag         string   `json:"tag"`
}
type WFAPIItem struct {
	Payload Payload `json:"payload"`
}
type Items struct {
	URLName  string `json:"url_name"`
	Thumb    string `json:"thumb"`
	ItemName string `json:"item_name"`
	ID       string `json:"id"`
	Vaulted  bool   `json:"vaulted,omitempty"`
}
type Payload struct {
	Items  []Items `json:"items"`
	Orders Orders  `json:"orders"`
}

type WFAPIItemsOrders struct {
	Payload Payload `json:"payload"`
	Include Include `json:"include"`
}
type User struct {
	IngameName string      `json:"ingame_name"`
	LastSeen   time.Time   `json:"last_seen"`
	Reputation int         `json:"reputation"`
	Region     string      `json:"region"`
	ID         string      `json:"id"`
	Avatar     interface{} `json:"avatar"`
	Status     string      `json:"status"`
}
type Orders []struct {
	OrderType    string    `json:"order_type"`
	LastUpdate   time.Time `json:"last_update"`
	Region       string    `json:"region"`
	Quantity     int       `json:"quantity"`
	Visible      bool      `json:"visible"`
	CreationDate time.Time `json:"creation_date"`
	Platinum     int       `json:"platinum"`
	Platform     string    `json:"platform"`
	User         User      `json:"user"`
	ID           string    `json:"id"`
	ModRank      int       `json:"mod_rank"`
}

func (a Orders) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a Orders) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a Orders) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[i].Platinum < a[j].Platinum
}

type En struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Ru struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Ko struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Fr struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Sv struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type De struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type ZhHant struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type ZhHans struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Pt struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Es struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type Pl struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type ItemsInSet struct {
	Icon           string   `json:"icon"`
	URLName        string   `json:"url_name"`
	SubIcon        string   `json:"sub_icon"`
	ModMaxRank     int      `json:"mod_max_rank"`
	Thumb          string   `json:"thumb"`
	SetRoot        bool     `json:"set_root"`
	QuantityForSet int      `json:"quantity_for_set,omitempty"`
	ID             string   `json:"id"`
	TradingTax     int      `json:"trading_tax"`
	Tags           []string `json:"tags"`
	MasteryLevel   int      `json:"mastery_level"`
	Ducats         int      `json:"ducats"`
	IconFormat     string   `json:"icon_format"`
	En             En       `json:"en"`
	Ru             Ru       `json:"ru"`
	Ko             Ko       `json:"ko"`
	Fr             Fr       `json:"fr"`
	Sv             Sv       `json:"sv"`
	De             De       `json:"de"`
	ZhHant         ZhHant   `json:"zh-hant"`
	ZhHans         ZhHans   `json:"zh-hans"`
	Pt             Pt       `json:"pt"`
	Es             Es       `json:"es"`
	Pl             Pl       `json:"pl"`
}
type Item struct {
	ID         string       `json:"id"`
	ItemsInSet []ItemsInSet `json:"items_in_set"`
}
type Include struct {
	Item Item `json:"item"`
}

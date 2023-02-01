package warframeapi

import "time"

type wfapi struct {
	Timestamp            time.Time            `json:"timestamp"`
	News                 []news               `json:"news"`
	Events               []events             `json:"events"`
	Alerts               []alerts             `json:"alerts"`
	Sortie               sortie               `json:"sortie"`
	SyndicateMissions    []syndicateMissions  `json:"syndicateMissions"`
	Fissures             []fissures           `json:"fissures"`
	GlobalUpgrades       []interface{}        `json:"globalUpgrades"`
	FlashSales           []flashSales         `json:"flashSales"`
	Invasions            []invasions          `json:"invasions"`
	DarkSectors          []interface{}        `json:"darkSectors"`
	VoidTrader           voidTrader           `json:"voidTrader"`
	DailyDeals           []dailyDeals         `json:"dailyDeals"`
	Simaris              simaris              `json:"simaris"`
	ConclaveChallenges   []conclaveChallenges `json:"conclaveChallenges"`
	PersistentEnemies    []interface{}        `json:"persistentEnemies"`
	EarthCycle           earthCycle           `json:"earthCycle"`
	CetusCycle           cetusCycle           `json:"cetusCycle"`
	CambionCycle         cambionCycle         `json:"cambionCycle"`
	ZarimanCycle         zarimanCycle         `json:"zarimanCycle"`
	WeeklyChallenges     []interface{}        `json:"weeklyChallenges"`
	ConstructionProgress constructionProgress `json:"constructionProgress"`
	VallisCycle          vallisCycle          `json:"vallisCycle"`
	Nightwave            nightwave            `json:"nightwave"`
	Kuva                 []interface{}        `json:"kuva"`
	Arbitration          arbitration          `json:"arbitration"`
	SentientOutposts     sentientOutposts     `json:"sentientOutposts"`
	SteelPath            steelPath            `json:"steelPath"`
	VaultTrader          vaultTrader          `json:"vaultTrader"`
}
type translations struct {
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
type news struct {
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
	Translations translations `json:"translations"`
	AsString     string       `json:"asString"`
}
type metadata struct {
}
type nextAlt struct {
	Expiry     time.Time `json:"expiry"`
	Activation time.Time `json:"activation"`
}
type events struct {
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
	Metadata          metadata      `json:"metadata"`
	CompletionBonuses []interface{} `json:"completionBonuses"`
	AltExpiry         time.Time     `json:"altExpiry"`
	AltActivation     time.Time     `json:"altActivation"`
	NextAlt           nextAlt       `json:"nextAlt"`
}
type variants struct {
	MissionType         string `json:"missionType"`
	Modifier            string `json:"modifier"`
	ModifierDescription string `json:"modifierDescription"`
	Node                string `json:"node"`
}
type sortie struct {
	ID          string     `json:"id"`
	Activation  time.Time  `json:"activation"`
	StartString string     `json:"startString"`
	Expiry      time.Time  `json:"expiry"`
	Active      bool       `json:"active"`
	RewardPool  string     `json:"rewardPool"`
	Variants    []variants `json:"variants"`
	Boss        string     `json:"boss"`
	Faction     string     `json:"faction"`
	Expired     bool       `json:"expired"`
	Eta         string     `json:"eta"`
}
type jobs struct {
	ID             string    `json:"id"`
	RewardPool     []string  `json:"rewardPool"`
	Type           string    `json:"type"`
	EnemyLevels    []int     `json:"enemyLevels"`
	StandingStages []int     `json:"standingStages"`
	MinMR          int       `json:"minMR"`
	Expiry         time.Time `json:"expiry"`
	TimeBound      string    `json:"timeBound,omitempty"`
}
type syndicateMissions struct {
	ID           string        `json:"id"`
	Activation   time.Time     `json:"activation"`
	StartString  string        `json:"startString"`
	Expiry       time.Time     `json:"expiry"`
	Active       bool          `json:"active"`
	Syndicate    string        `json:"syndicate"`
	SyndicateKey string        `json:"syndicateKey"`
	Nodes        []interface{} `json:"nodes"`
	Jobs         []jobs        `json:"jobs"`
	Eta          string        `json:"eta"`
}
type fissures struct {
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
type flashSales struct {
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
type countedItems struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
	Key   string `json:"key"`
}
type attackerReward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []countedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type reward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []countedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type attacker struct {
	Reward     reward `json:"reward"`
	Faction    string `json:"faction"`
	FactionKey string `json:"factionKey"`
}
type defenderReward struct {
	Items        []interface{}  `json:"items"`
	CountedItems []countedItems `json:"countedItems"`
	Credits      int            `json:"credits"`
	AsString     string         `json:"asString"`
	ItemString   string         `json:"itemString"`
	Thumbnail    string         `json:"thumbnail"`
	Color        int            `json:"color"`
}
type defender struct {
	Reward     reward `json:"reward"`
	Faction    string `json:"faction"`
	FactionKey string `json:"factionKey"`
}
type invasions struct {
	ID               string         `json:"id"`
	Activation       time.Time      `json:"activation"`
	StartString      string         `json:"startString"`
	Node             string         `json:"node"`
	NodeKey          string         `json:"nodeKey"`
	Desc             string         `json:"desc"`
	AttackerReward   attackerReward `json:"attackerReward"`
	AttackingFaction string         `json:"attackingFaction"`
	Attacker         attacker       `json:"attacker"`
	DefenderReward   defenderReward `json:"defenderReward"`
	DefendingFaction string         `json:"defendingFaction"`
	Defender         defender       `json:"defender"`
	VsInfestation    bool           `json:"vsInfestation"`
	Count            int            `json:"count"`
	RequiredRuns     int            `json:"requiredRuns"`
	Completion       float64        `json:"completion"`
	Completed        bool           `json:"completed"`
	Eta              string         `json:"eta"`
	RewardTypes      []string       `json:"rewardTypes"`
}
type voidTrader struct {
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
type dailyDeals struct {
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
type simaris struct {
	Target         string `json:"target"`
	IsTargetActive bool   `json:"isTargetActive"`
	AsString       string `json:"asString"`
}
type conclaveChallenges struct {
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
type earthCycle struct {
	ID         string    `json:"id"`
	Expiry     time.Time `json:"expiry"`
	Activation time.Time `json:"activation"`
	IsDay      bool      `json:"isDay"`
	State      string    `json:"state"`
	TimeLeft   string    `json:"timeLeft"`
}
type cetusCycle struct {
	ID          string    `json:"id"`
	Expiry      time.Time `json:"expiry"`
	Activation  time.Time `json:"activation"`
	IsDay       bool      `json:"isDay"`
	State       string    `json:"state"`
	TimeLeft    string    `json:"timeLeft"`
	IsCetus     bool      `json:"isCetus"`
	ShortString string    `json:"shortString"`
}
type cambionCycle struct {
	ID         string    `json:"id"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
	TimeLeft   string    `json:"timeLeft"`
	Active     string    `json:"active"`
}
type zarimanCycle struct {
	ID              string    `json:"id"`
	BountiesEndDate time.Time `json:"bountiesEndDate"`
	Expiry          time.Time `json:"expiry"`
	Activation      time.Time `json:"activation"`
	IsCorpus        bool      `json:"isCorpus"`
	State           string    `json:"state"`
	TimeLeft        string    `json:"timeLeft"`
	ShortString     string    `json:"shortString"`
}
type constructionProgress struct {
	ID                string `json:"id"`
	FomorianProgress  string `json:"fomorianProgress"`
	RazorbackProgress string `json:"razorbackProgress"`
	UnknownProgress   string `json:"unknownProgress"`
}
type vallisCycle struct {
	ID          string    `json:"id"`
	Expiry      time.Time `json:"expiry"`
	IsWarm      bool      `json:"isWarm"`
	State       string    `json:"state"`
	Activation  time.Time `json:"activation"`
	TimeLeft    string    `json:"timeLeft"`
	ShortString string    `json:"shortString"`
}
type params struct {
}
type activeChallenges struct {
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
type nightwave struct {
	ID                 string             `json:"id"`
	Activation         time.Time          `json:"activation"`
	StartString        string             `json:"startString"`
	Expiry             time.Time          `json:"expiry"`
	Active             bool               `json:"active"`
	Season             int                `json:"season"`
	Tag                string             `json:"tag"`
	Phase              int                `json:"phase"`
	Params             params             `json:"params"`
	PossibleChallenges []interface{}      `json:"possibleChallenges"`
	ActiveChallenges   []activeChallenges `json:"activeChallenges"`
	RewardTypes        []string           `json:"rewardTypes"`
}
type arbitration struct {
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
type mission struct {
	Node    string `json:"node"`
	Faction string `json:"faction"`
	Type    string `json:"type"`
}
type sentientOutposts struct {
	Mission    mission   `json:"mission"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
	Active     bool      `json:"active"`
	ID         string    `json:"id"`
}
type currentReward struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type rotation struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type evergreens struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}
type incursions struct {
	ID         string    `json:"id"`
	Activation time.Time `json:"activation"`
	Expiry     time.Time `json:"expiry"`
}
type steelPath struct {
	CurrentReward currentReward `json:"currentReward"`
	Activation    time.Time     `json:"activation"`
	Expiry        time.Time     `json:"expiry"`
	Remaining     string        `json:"remaining"`
	Rotation      []rotation    `json:"rotation"`
	Evergreens    []evergreens  `json:"evergreens"`
	Incursions    incursions    `json:"incursions"`
}
type inventory struct {
	Item    string      `json:"item"`
	Ducats  int         `json:"ducats"`
	Credits interface{} `json:"credits"`
}
type schedule struct {
	Expiry time.Time `json:"expiry"`
	Item   string    `json:"item"`
}
type vaultTrader struct {
	ID           string      `json:"id"`
	Activation   time.Time   `json:"activation"`
	StartString  string      `json:"startString"`
	Expiry       time.Time   `json:"expiry"`
	Active       bool        `json:"active"`
	Character    string      `json:"character"`
	Location     string      `json:"location"`
	Inventory    []inventory `json:"inventory"`
	PsID         string      `json:"psId"`
	EndString    string      `json:"endString"`
	InitialStart time.Time   `json:"initialStart"`
	Completed    bool        `json:"completed"`
	Schedule     []schedule  `json:"schedule"`
}
type alerts struct {
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
type wfAPIItem struct {
	Payload payload `json:"payload"`
}
type items struct {
	URLName  string `json:"url_name"`
	Thumb    string `json:"thumb"`
	ItemName string `json:"item_name"`
	ID       string `json:"id"`
	Vaulted  bool   `json:"vaulted,omitempty"`
}
type payload struct {
	Items  []items `json:"items"`
	Orders orders  `json:"orders"`
}

type wfAPIItemsOrders struct {
	Payload payload `json:"payload"`
	Include include `json:"include"`
}
type user struct {
	IngameName string      `json:"ingame_name"`
	LastSeen   time.Time   `json:"last_seen"`
	Reputation int         `json:"reputation"`
	Region     string      `json:"region"`
	ID         string      `json:"id"`
	Avatar     interface{} `json:"avatar"`
	Status     string      `json:"status"`
}
type orders []struct {
	OrderType    string    `json:"order_type"`
	LastUpdate   time.Time `json:"last_update"`
	Region       string    `json:"region"`
	Quantity     int       `json:"quantity"`
	Visible      bool      `json:"visible"`
	CreationDate time.Time `json:"creation_date"`
	Platinum     int       `json:"platinum"`
	Platform     string    `json:"platform"`
	User         user      `json:"user"`
	ID           string    `json:"id"`
	ModRank      int       `json:"mod_rank"`
}

func (a orders) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a orders) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a orders) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[i].Platinum < a[j].Platinum
}

type en struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type ru struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type ko struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type fr struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type sv struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type de struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type zhHant struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type zhHans struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type pt struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type es struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type pl struct {
	ItemName    string        `json:"item_name"`
	Description string        `json:"description"`
	WikiLink    string        `json:"wiki_link"`
	Drop        []interface{} `json:"drop"`
}
type itemsInSet struct {
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
	En             en       `json:"en"`
	Ru             ru       `json:"ru"`
	Ko             ko       `json:"ko"`
	Fr             fr       `json:"fr"`
	Sv             sv       `json:"sv"`
	De             de       `json:"de"`
	ZhHant         zhHant   `json:"zh-hant"`
	ZhHans         zhHans   `json:"zh-hans"`
	Pt             pt       `json:"pt"`
	Es             es       `json:"es"`
	Pl             pl       `json:"pl"`
}
type item struct {
	ID         string       `json:"id"`
	ItemsInSet []itemsInSet `json:"items_in_set"`
}
type include struct {
	Item item `json:"item"`
}

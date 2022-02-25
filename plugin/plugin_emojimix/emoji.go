// Package emojimix 合成emoji
package emojimix

var emojis = map[rune]int64{
	128516: 20201001, // 😄 grinning face with smiling eyes
	128512: 20201001, // 😀 grinning face
	128578: 20201001, // 🙂 slightly smiling face
	128579: 20201001, // 🙃 upside-down face
	128521: 20201001, // 😉 winking face
	128522: 20201001, // 😊 smiling face with smiling eyes
	128518: 20201001, // 😆 grinning squinting face
	128515: 20201001, // 😃 grinning face with big eyes
	128513: 20201001, // 😁 beaming face with smiling eyes
	129315: 20201001, // 🤣 rolling on the floor laughing
	128517: 20201001, // 😅 grinning face with sweat
	128514: 20201001, // 😂 face with tears of joy
	128519: 20201001, // 😇 smiling face with halo
	129392: 20201001, // 🥰 smiling face with hearts
	128525: 20201001, // 😍 smiling face with heart-eyes
	128536: 20201001, // 😘 face blowing a kiss
	129321: 20201001, // 🤩 star-struck
	128535: 20201001, // 😗 kissing face
	128538: 20201001, // 😚 kissing face with closed eyes
	128537: 20201001, // 😙 kissing face with smiling eyes
	128539: 20201001, // 😛 face with tongue
	128541: 20201001, // 😝 squinting face with tongue
	128523: 20201001, // 😋 face savoring food
	129394: 20201001, // 🥲 smiling face with tear
	129297: 20201001, // 🤑 money-mouth face
	128540: 20201001, // 😜 winking face with tongue
	129303: 20201001, // 🤗 smiling face with open hands hugs
	129323: 20201001, // 🤫 shushing face quiet whisper
	129300: 20201001, // 🤔 thinking face question hmmm
	129325: 20201001, // 🤭 face with hand over mouth embarrassed
	129320: 20201001, // 🤨 face with raised eyebrow question
	129296: 20201001, // 🤐 zipper-mouth face
	128528: 20201001, // 😐 neutral face
	128529: 20201001, // 😑 expressionless face
	128566: 20201001, // 😶 face without mouth
	129322: 20201001, // 🤪 zany face
	128527: 20201001, // 😏 smirking face suspicious
	128530: 20201001, // 😒 unamused face
	128580: 20201001, // 🙄 face with rolling eyes
	128556: 20201001, // 😬 grimacing face
	128558: 20210218, // 😮 face exhaling
	129317: 20201001, // 🤥 lying face
	128524: 20201001, // 😌 relieved face
	128532: 20201001, // 😔 pensive face
	128554: 20201001, // 😪 sleepy face
	129316: 20201001, // 🤤 drooling face
	128564: 20201001, // 😴 sleeping face
	128567: 20201001, // 😷 face with medical mask
	129298: 20201001, // 🤒 face with thermometer
	129301: 20201001, // 🤕 face with head-bandage
	129314: 20201001, // 🤢 nauseated face
	129326: 20201001, // 🤮 face vomiting throw
	129319: 20201001, // 🤧 sneezing face
	129397: 20201001, // 🥵 hot face warm
	129398: 20201001, // 🥶 cold face freezing ice
	128565: 20201001, // 😵 face with crossed-out eyes
	129396: 20201001, // 🥴 woozy face drunk tipsy drug high
	129327: 20201001, // 🤯 exploding head mindblow
	129312: 20201001, // 🤠 cowboy hat face
	129395: 20201001, // 🥳 partying face
	129400: 20201001, // 🥸 disguised face
	129488: 20201001, // 🧐 face with monocle glasses
	128526: 20201001, // 😎 smiling face with sunglasses
	128533: 20201001, // 😕 confused face
	128543: 20201001, // 😟 worried face
	128577: 20201001, // 🙁 slightly frowning face
	128559: 20201001, // 😯 hushed face
	128562: 20201001, // 😲 astonished face
	129299: 20201001, // 🤓 nerd face glasses
	128563: 20201001, // 😳 flushed face
	129402: 20201001, // 🥺 pleading face
	128551: 20201001, // 😧 anguished face
	128552: 20201001, // 😨 fearful face
	128550: 20201001, // 😦 frowning face with open mouth
	128560: 20201001, // 😰 anxious face with sweat
	128549: 20201001, // 😥 sad but relieved face
	128557: 20201001, // 😭 loudly crying face
	128553: 20201001, // 😩 weary face
	128546: 20201001, // 😢 crying face
	128547: 20201001, // 😣 persevering face
	128544: 20201001, // 😠 angry face
	128531: 20201001, // 😓 downcast face with sweat
	128534: 20201001, // 😖 confounded face
	129324: 20201001, // 🤬 face with symbols on mouth
	128542: 20201001, // 😞 disappointed face
	128555: 20201001, // 😫 tired face
	128548: 20201001, // 😤 face with steam from nose
	129393: 20201001, // 🥱 yawning face
	128169: 20201001, // 💩 pile of poo
	128545: 20201001, // 😡 pouting face
	128561: 20201001, // 😱 face screaming in fear
	128127: 20201001, // 👿 angry face with horns
	128128: 20201001, // 💀 skull
	128125: 20201001, // 👽 alien
	128520: 20201001, // 😈 smiling face with horns devil
	129313: 20201001, // 🤡 clown face
	128123: 20201001, // 👻 ghost
	129302: 20201001, // 🤖 robot
	128175: 20201001, // 💯 hundred points percent
	128064: 20201001, // 👀 eyes
	127801: 20201001, // 🌹 rose flower
	127804: 20201001, // 🌼 blossom flower
	127799: 20201001, // 🌷 tulip flower
	127797: 20201001, // 🌵 cactus
	127821: 20201001, // 🍍 pineapple
	127874: 20201001, // 🎂 birthday cake
	127751: 20210831, // 🌇 sunset
	129473: 20201001, // 🧁 cupcake muffin
	127911: 20210521, // 🎧 headphone earphone
	127800: 20210218, // 🌸 cherry blossom flower
	129440: 20201001, // 🦠 microbe germ bacteria virus covid corona
	128144: 20201001, // 💐 bouquet flowers
	127789: 20201001, // 🌭 hot dog food
	128139: 20201001, // 💋 kiss mark lips
	127875: 20201001, // 🎃 jack-o-lantern pumpkin
	129472: 20201001, // 🧀 cheese wedge
	9749:   20201001, // ☕ hot beverage coffee cup tea
	127882: 20201001, // 🎊 confetti ball
	127880: 20201001, // 🎈 balloon
	9924:   20201001, // ⛄ snowman without snow
	128142: 20201001, // 💎 gem stone crystal diamond
	127794: 20201001, // 🌲 evergreen tree
	129410: 20210218, // 🦂 scorpion
	128584: 20201001, // 🙈 see-no-evil monkey
	128148: 20201001, // 💔 broken heart
	128140: 20201001, // 💌 love letter heart
	128152: 20201001, // 💘 heart with arrow
	128159: 20201001, // 💟 heart decoration
	128158: 20201001, // 💞 revolving hearts
	128147: 20201001, // 💓 beating heart
	128149: 20201001, // 💕 two hearts
	128151: 20201001, // 💗 growing heart
	129505: 20201001, // 🧡 orange heart
	128155: 20201001, // 💛 yellow heart
	10084:  20210218, // ❤ mending heart
	128156: 20201001, // 💜 purple heart
	128154: 20201001, // 💚 green heart
	128153: 20201001, // 💙 blue heart
	129294: 20201001, // 🤎 brown heart
	129293: 20201001, // 🤍 white heart
	128420: 20201001, // 🖤 black heart
	128150: 20201001, // 💖 sparkling heart
	128157: 20201001, // 💝 heart with ribbon
	127873: 20211115, // 🎁 wrapped-gift
	129717: 20211115, // 🪵 wood
	127942: 20211115, // 🏆 trophy
	127838: 20210831, // 🍞 bread
	128240: 20201001, // 📰 newspaper
	128302: 20201001, // 🔮 crystal ball
	128081: 20201001, // 👑 crown
	128055: 20201001, // 🐷 pig face
	129412: 20210831, // 🦄 unicorn
	127771: 20201001, // 🌛 first quarter moon face
	129420: 20201001, // 🦌 deer
	129668: 20210521, // 🪄 magic wand
	128171: 20201001, // 💫 dizzy
	128049: 20201001, // 🐱 meow cat face
	129409: 20201001, // 🦁 lion
	128293: 20201001, // 🔥 fire
	128038: 20210831, // 🐦 bird
	129415: 20201001, // 🦇 bat
	129417: 20210831, // 🦉 owl
	127752: 20201001, // 🌈 rainbow
	128053: 20201001, // 🐵 monkey face
	128029: 20201001, // 🐝 honeybee bumblebee wasp
	128034: 20201001, // 🐢 turtle
	128025: 20201001, // 🐙 octopus
	129433: 20201001, // 🦙 llama alpaca
	128016: 20210831, // 🐐 goat
	128060: 20201001, // 🐼 panda
	128040: 20201001, // 🐨 koala
	129445: 20201001, // 🦥 sloth
	128059: 20210831, // 🐻 bear
	128048: 20201001, // 🐰 rabbit face
	129428: 20201001, // 🦔 hedgehog
	128054: 20211115, // 🐶 dog puppy
	128041: 20211115, // 🐩 poodle dog
	129437: 20211115, // 🦝 raccoon
	128039: 20211115, // 🐧 penguin
	128012: 20210218, // 🐌 snail
	128045: 20201001, // 🐭 mouse face rat
	128031: 20210831, // 🐟 fish
	127757: 20201001, // 🌍 globe showing Europe-Africa
	127774: 20201001, // 🌞 sun with face
	127775: 20201001, // 🌟 glowing star
	11088:  20201001, // ⭐ star
	127772: 20201001, // 🌜 last quarter moon face
	129361: 20201001, // 🥑 avocado
	127820: 20211115, // 🍌 banana
	127827: 20210831, // 🍓 strawberry
	127819: 20210521, // 🍋 lemon
	127818: 20211115, // 🍊 tangerine orange
}

var qqface = map[int]rune{
	0:   128558, // 😮 face exhaling
	1:   128556, // 😬 grimacing face
	2:   128525, // 😍 smiling face with heart-eyes
	4:   128526, // 😎 smiling face with sunglasses
	5:   128557, // 😭 loudly crying face
	6:   129402, // 🥺 pleading face
	7:   129296, // 🤐 zipper-mouth face
	8:   128554, // 😪 sleepy face
	11:  128545, // 😡 pouting face
	12:  128539, // 😛 face with tongue
	13:  128513, // 😁 beaming face with smiling eyes
	14:  128578, // 🙂 slightly smiling face
	15:  128577, // 🙁 slightly frowning face
	16:  128526, // 😎 smiling face with sunglasses
	19:  129326, // 🤮 face vomiting throw
	20:  129325, // 🤭 face with hand over mouth embarrassed
	21:  128522, // 😊 smiling face with smiling eyes
	23:  128533, // 😕 confused face
	24:  128523, // 😋 face savoring food
	27:  128531, // 😓 downcast face with sweat
	28:  128516, // 😄 grinning face with smiling eyes
	31:  129324, // 🤬 face with symbols on mouth
	32:  129300, // 🤔 thinking face question hmmm
	33:  129323, // 🤫 shushing face quiet whisper
	34:  128565, // 😵 face with crossed-out eyes
	35:  128547, // 😣 persevering face
	37:  128128, // 💀 skull
	46:  128055, // 🐷 pig face
	53:  127874, // 🎂 birthday cake
	59:  128169, // 💩 pile of poo
	60:  9749,   // ☕ hot beverage coffee cup tea
	63:  127801, // 🌹 rose flower
	66:  10084,  // ❤ mending heart
	67:  128148, // 💔 broken heart
	69:  127873, // 🎁 wrapped-gift
	74:  127774, // 🌞 sun with face
	75:  127772, // 🌜 last quarter moon face
	96:  128517, // 😅 grinning face with sweat
	104: 129393, // 🥱 yawning face
	109: 128535, // 😗 kissing face
	110: 128562, // 😲 astonished face
	111: 129402, // 🥺 pleading face
	172: 128539, // 😛 face with tongue
	182: 128514, // 😂 face with tears of joy
	187: 128123, // 👻 ghost
	247: 128567, // 😷 face with medical mask
	272: 128579, // 🙃 upside-down face
	320: 129395, // 🥳 partying face
	325: 128561, // 😱 face screaming in fear
}

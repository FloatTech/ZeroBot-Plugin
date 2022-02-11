// Package emojimix 合成emoji
package emojimix

var emojis = map[string]int64{
	"😄": 20201001, // grinning face with smiling eyes
	"😀": 20201001, // grinning face
	"🙂": 20201001, // slightly smiling face
	"🙃": 20201001, // upside-down face
	"😉": 20201001, // winking face
	"😊": 20201001, // smiling face with smiling eyes
	"😆": 20201001, // grinning squinting face
	"😃": 20201001, // grinning face with big eyes
	"😁": 20201001, // beaming face with smiling eyes
	"🤣": 20201001, // rolling on the floor laughing
	"😅": 20201001, // grinning face with sweat
	"😂": 20201001, // face with tears of joy
	"😇": 20201001, // smiling face with halo
	"🥰": 20201001, // smiling face with hearts
	"😍": 20201001, // smiling face with heart-eyes
	"😘": 20201001, // face blowing a kiss
	"🤩": 20201001, // star-struck
	"😗": 20201001, // kissing face
	"😚": 20201001, // kissing face with closed eyes
	"😙": 20201001, // kissing face with smiling eyes
	"😛": 20201001, // face with tongue
	"😝": 20201001, // squinting face with tongue
	"😋": 20201001, // face savoring food
	"🥲": 20201001, // smiling face with tear
	"🤑": 20201001, // money-mouth face
	"😜": 20201001, // winking face with tongue
	"🤗": 20201001, // smiling face with open hands hugs
	"🤫": 20201001, // shushing face quiet whisper
	"🤔": 20201001, // thinking face question hmmm
	"🤭": 20201001, // face with hand over mouth embarrassed
	"🤨": 20201001, // face with raised eyebrow question
	"🤐": 20201001, // zipper-mouth face
	"😐": 20201001, // neutral face
	"😑": 20201001, // expressionless face
	"😶": 20201001, // face without mouth
	"🤪": 20201001, // zany face
	"😏": 20201001, // smirking face suspicious
	"😒": 20201001, // unamused face
	"🙄": 20201001, // face with rolling eyes
	"😬": 20201001, // grimacing face
	"🤥": 20201001, // lying face
	"😌": 20201001, // relieved face
	"😔": 20201001, // pensive face
	"😪": 20201001, // sleepy face
	"🤤": 20201001, // drooling face
	"😴": 20201001, // sleeping face
	"😷": 20201001, // face with medical mask
	"🤒": 20201001, // face with thermometer
	"🤕": 20201001, // face with head-bandage
	"🤢": 20201001, // nauseated face
	"🤮": 20201001, // face vomiting throw
	"🤧": 20201001, // sneezing face
	"🥵": 20201001, // hot face warm
	"🥶": 20201001, // cold face freezing ice
	"😵": 20201001, // face with crossed-out eyes
	"🥴": 20201001, // woozy face drunk tipsy drug high
	"🤯": 20201001, // exploding head mindblow
	"🤠": 20201001, // cowboy hat face
	"🥳": 20201001, // partying face
	"🥸": 20201001, // disguised face
	"🧐": 20201001, // face with monocle glasses
	"😎": 20201001, // smiling face with sunglasses
	"😕": 20201001, // confused face
	"😟": 20201001, // worried face
	"🙁": 20201001, // slightly frowning face
	"😮": 20201001, // face with open mouth
	"😯": 20201001, // hushed face
	"😲": 20201001, // astonished face
	"🤓": 20201001, // nerd face glasses
	"😳": 20201001, // flushed face
	"🥺": 20201001, // pleading face
	"😧": 20201001, // anguished face
	"😨": 20201001, // fearful face
	"😦": 20201001, // frowning face with open mouth
	"😰": 20201001, // anxious face with sweat
	"😥": 20201001, // sad but relieved face
	"😭": 20201001, // loudly crying face
	"😩": 20201001, // weary face
	"😢": 20201001, // crying face
	"😣": 20201001, // persevering face
	"😠": 20201001, // angry face
	"😓": 20201001, // downcast face with sweat
	"😖": 20201001, // confounded face
	"🤬": 20201001, // face with symbols on mouth
	"😞": 20201001, // disappointed face
	"😫": 20201001, // tired face
	"😤": 20201001, // face with steam from nose
	"🥱": 20201001, // yawning face
	"💩": 20201001, // pile of poo
	"😡": 20201001, // pouting face
	"😱": 20201001, // face screaming in fear
	"👿": 20201001, // angry face with horns
	"💀": 20201001, // skull
	"👽": 20201001, // alien
	"😈": 20201001, // smiling face with horns devil
	"🤡": 20201001, // clown face
	"👻": 20201001, // ghost
	"🤖": 20201001, // robot
	"💯": 20201001, // hundred points percent
	"👀": 20201001, // eyes
	"🌹": 20201001, // rose flower
	"🌼": 20201001, // blossom flower
	"🌷": 20201001, // tulip flower
	"🌵": 20201001, // cactus
	"🍍": 20201001, // pineapple
	"🎂": 20201001, // birthday cake
	"🌇": 20210831, // sunset
	"🧁": 20201001, // cupcake muffin
	"🎧": 20210521, // headphone earphone
	"🌸": 20210218, // cherry blossom flower
	"🦠": 20201001, // microbe germ bacteria virus covid corona
	"💐": 20201001, // bouquet flowers
	"🌭": 20201001, // hot dog food
	"💋": 20201001, // kiss mark lips
	"🎃": 20201001, // jack-o-lantern pumpkin
	"🧀": 20201001, // cheese wedge
	"☕": 20201001, // hot beverage coffee cup tea
	"🎊": 20201001, // confetti ball
	"🎈": 20201001, // balloon
	"⛄": 20201001, // snowman without snow
	"💎": 20201001, // gem stone crystal diamond
	"🌲": 20201001, // evergreen tree
	"🦂": 20210218, // scorpion
	"🙈": 20201001, // see-no-evil monkey
	"💔": 20201001, // broken heart
	"💌": 20201001, // love letter heart
	"💘": 20201001, // heart with arrow
	"💟": 20201001, // heart decoration
	"💞": 20201001, // revolving hearts
	"💓": 20201001, // beating heart
	"💕": 20201001, // two hearts
	"💗": 20201001, // growing heart
	"🧡": 20201001, // orange heart
	"💛": 20201001, // yellow heart
	"💜": 20201001, // purple heart
	"💚": 20201001, // green heart
	"💙": 20201001, // blue heart
	"🤎": 20201001, // brown heart
	"🤍": 20201001, // white heart
	"🖤": 20201001, // black heart
	"💖": 20201001, // sparkling heart
	"💝": 20201001, // heart with ribbon
	"🎁": 20211115, // wrapped-gift
	"🪵": 20211115, // wood
	"🏆": 20211115, // trophy
	"🍞": 20210831, // bread
	"📰": 20201001, // newspaper
	"🔮": 20201001, // crystal ball
	"👑": 20201001, // crown
	"🐷": 20201001, // pig face
	"🦄": 20210831, // unicorn
	"🌛": 20201001, // first quarter moon face
	"🦌": 20201001, // deer
	"🪄": 20210521, // magic wand
	"💫": 20201001, // dizzy
	"🐱": 20201001, // meow cat face
	"🦁": 20201001, // lion
	"🔥": 20201001, // fire
	"🐦": 20210831, // bird
	"🦇": 20201001, // bat
	"🦉": 20210831, // owl
	"🌈": 20201001, // rainbow
	"🐵": 20201001, // monkey face
	"🐝": 20201001, // honeybee bumblebee wasp
	"🐢": 20201001, // turtle
	"🐙": 20201001, // octopus
	"🦙": 20201001, // llama alpaca
	"🐐": 20210831, // goat
	"🐼": 20201001, // panda
	"🐨": 20201001, // koala
	"🦥": 20201001, // sloth
	"🐻": 20210831, // bear
	"🐰": 20201001, // rabbit face
	"🦔": 20201001, // hedgehog
	"🐶": 20211115, // dog puppy
	"🐩": 20211115, // poodle dog
	"🦝": 20211115, // raccoon
	"🐧": 20211115, // penguin
	"🐌": 20210218, // snail
	"🐭": 20201001, // mouse face rat
	"🐟": 20210831, // fish
	"🌍": 20201001, // globe showing Europe-Africa
	"🌞": 20201001, // sun with face
	"🌟": 20201001, // glowing star
	"⭐": 20201001, // star
	"🌜": 20201001, // last quarter moon face
	"🥑": 20201001, // avocado
	"🍌": 20211115, // banana
	"🍓": 20210831, // strawberry
	"🍋": 20210521, // lemon
	"🍊": 20211115, // tangerine orange
}

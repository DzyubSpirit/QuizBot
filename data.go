package main

var BibleBooksEn = []string{"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy", "Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel", "1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles", "Ezra", "Nehemiah", "Esther", "Job", "Psalms", "Proverbs", "Ecclesiastes", "Song of Songs", "Isaiah", "Jeremiah", "Lamentations", "Ezekiel", "Daniel", "Hosea", "Joel", "Amos", "Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk", "Zephaniah", "Haggai", "Zechariah", "Malachi", "Matthew", "Mark", "Luke", "John", "Acts", "Romans", "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians", "Philippians", "Colossians", "1 Thessalonians", "2 Thessalonians", "1 Timothy", "2 Timothy", "Titus", "Philemon", "Hebrews", "James", "1 Peter", "2 Peter", "1 John", "2 John", "3 John", "Jude", "Revelation"}

var DefaultTopic = "books_en"
var QuestionableMap = map[string]Questionable{
	"books_ru": BookAfterQuestion{
		Books:         []string{"Бытие", "Исход", "Левит", "Числа", "Второзаконие", "Иисус Навин", "Судьей", "Руфь", "1 Царств", "2 Царств", "3 Царств", "4 Царств", "1 Паралипоменон", "2 Паралипоменон", "Ездра", "Неемия", "Есфирь", "Иов", "Псалтырь", "Притчи", "Екклесиаст", "Книга Песнь Песней", "Исаия", "Иеремия", "Плач Иеремии", "Иезекииль", "Даниил", "Осия", "Иоиль", "Амос", "Авдий", "Иона", "Михей", "Наум", "Аввакум", "Софония", "Аггей", "Захария", "Малахия", "Матфея", "Марка", "Луки", "Иоанна", "Деяния", "Иакова", "1 Петра", "2 Петра", "1 Иоанна", "2 Иоанна", "3 Иоанна", "Иуды", "Римлянам", "1 Коринфянам", "2 Коринфянам", "Галатам", "Ефесянам", "Филиппийцам", "Колоссянам", "1 Фессалоникийцам", "2 Фессалоникийцам", "1 Тимофею", "2 Тимофею", "Титу", "Филимону", "Евреям", "Откровение"},
		Question:      "Какая книга идет после ",
		AnswersNumber: 5,
	},
	"books_en": BookAfterQuestion{
		Books:         BibleBooksEn,
		Question:      "What book is after ",
		AnswersNumber: 5,
	},
	"before_after_en": BeforeAfterQuestion{
		Books: BibleBooksEn,
		Question: "Where is '%s' relative to '%s'?",
	},
}
var Topics TopicsMap

func init() {
	Topics = make(TopicsMap, len(QuestionableMap))
	for k, q := range QuestionableMap {
		Topics[k] = Topic{Questions: q.ToQuestions()}
	}
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/telebot.v3"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Get the bot token from the environment variables
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN is not set in the environment variables")
	}

	pref := tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Handle new users joining the group
	bot.Handle(tb.OnUserJoined, func(c tb.Context) error {
		handleUserJoined(bot, c)
		return nil
	})

	// Handle /start command
	bot.Handle("/start", func(c tb.Context) error {
		handleStartCommand(bot, c)
		return nil
	})

	// YUP
	fmt.Println("Telegram Bot, Start listening ...")

	bot.Start()
}

// handleUserJoined handles the event when a user joins the group.
func handleUserJoined(bot *tb.Bot, c tb.Context) {
	user := c.Sender()
	chat := c.Chat()

	// Inform the group about the new user
	groupNotification := fmt.Sprintf("مرحباً %s أنت على وشك الإنضمام إلينا. يرجى التحقق من رسائلك الخاصة وإكمال عملية التحقق خلال 120 ثانية.", user.FirstName)
	bot.Send(chat, groupNotification)

	// Send private message to the user
	privateMessage := fmt.Sprintf("مرحباً! أنت على وشك الانضمام إلى المجموعة %s، ولكن قبل ذلك المرجو إكمال عملية التحقق أولا.", chat.Title)
	bot.Send(user, privateMessage)

	// Provide options
	options := "اختر السبب الذي يجعلك تنضم إلى المجموعة:\n" +
		"1. للتعلم المزيد عن لغة Go\n" +
		"2. لمشاركة معرفتي مع مجتمع مطوري Go\n" +
		"3. للتسلية والترفيه\n" +
		"4. لنشر إعلاناتي"
	bot.Send(user, options)

	// Create a channel to receive the user's response
	answerChan := make(chan string)

	// Listen for the user's response
	go func() {
		bot.Handle(tb.OnText, func(c tb.Context) error {
			answerChan <- c.Message().Text
			return nil
		})
	}()

	// Wait for the user's answer or timeout after 20 seconds
	var chosenOption string
	select {
	case chosenOption = <-answerChan:
		if chosenOption == "1" || chosenOption == "2" {
			// Proceed to math problem if a valid option is chosen
			if askMathProblem(bot, user) {
				// If math problem solved correctly, welcome the user to the group
				welcomeUserToGroup(bot, chat, user)
			} else {
				bot.Send(user, "إجابة غير صحيحة! سيتم رفض طلبك للانضمام.")
				bot.DeclineJoinRequest(chat, user)
			}
		} else {
			bot.Send(user, "تم اختيار إجابة غير صحيحة! سيتم رفض طلبك للانضمام.")
			bot.DeclineJoinRequest(chat, user)
		}
	case <-time.After(120 * time.Second):
		bot.Send(user, "لم تقم بتحديد أي خيار! سيتم رفض طلبك للانضمام.")
		bot.DeclineJoinRequest(chat, user)
	}
}

// askMathProblem asks the user to solve a math problem and returns true if they solve it correctly.
func askMathProblem(bot *tb.Bot, user *tb.User) bool {
	// Generate a random math problem
	num1 := rand.Intn(10)
	num2 := rand.Intn(10)
	correctAnswer := num1 + num2

	// Send the math problem to the user in Arabic
	problem := fmt.Sprintf("يرجى حل هذه المسألة خلال 15 ثانية: %d + %d = ?", num1, num2)
	bot.Send(user, problem)

	// Create a channel to receive the user's math answer
	mathAnswerChan := make(chan string)

	// Listen for the user's response
	go func() {
		bot.Handle(tb.OnText, func(c tb.Context) error {
			mathAnswerChan <- c.Message().Text
			return nil
		})
	}()

	// Wait for the user's math answer or timeout after 6 seconds
	select {
	case mathAnswer := <-mathAnswerChan:
		if userMathAnswer, err := strconv.Atoi(mathAnswer); err == nil && userMathAnswer == correctAnswer {
			bot.Send(user, "إجابة صحيحة! يمكنك الآن الانضمام إلى المجموعة.")
			return true
		} else {
			bot.Send(user, "إجابة غير صحيحة! سيتم رفض طلبك للانضمام.")
			return false
		}
	case <-time.After(15 * time.Second):
		bot.Send(user, "انتهى الوقت! سيتم رفض طلبك للانضمام.")
		return false
	}
}

// welcomeUserToGroup sends a welcome message and adds the user to the group.
func welcomeUserToGroup(bot *tb.Bot, chat *tb.Chat, user *tb.User) {
	welcomeMessage := fmt.Sprintf("سادتي وسيداتي رحبوا معنا بالوافد الجديد %s! لقد تم قبوله معنا 🤠🎉 ", user.FirstName)
	bot.Send(chat, welcomeMessage)
}

// handleStartCommand handles the /start command and sends a welcome message.
func handleStartCommand(bot *tb.Bot, c tb.Context) {
	startMessage := "مرحباً! أنا البواب الحارس، الغوفر 🍉، أساعدكم على طرد الوافدين الجدد المخادعين. "
	bot.Send(c.Chat(), startMessage)
}

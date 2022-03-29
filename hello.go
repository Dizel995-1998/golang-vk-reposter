package main

import (
	"PostsGenerator/VkParser"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"os"
	"strings"
	"time"
)

// Функция публикация поста из ВК в Телеграмм
func publishVkPostToTelegram(telegramChatId int64, post VkParser.VkPost, tgBot *tgbotapi.BotAPI) (err error) {
	photoContent := make([]interface{}, 0)
	preparedMsg := post.GetText() + " " + strings.Join(post.GetVideoLinks(), " ")

	isFirstElement := false
	for _, photo := range post.GetPictureLinks() {
		inputMediaPhoto := tgbotapi.NewInputMediaPhoto(photo)

		if isFirstElement == false {
			inputMediaPhoto.Caption = preparedMsg
			isFirstElement = true
		}

		photoContent = append(photoContent, inputMediaPhoto)
	}

	// Если у ВК поста есть фотографии - отправляем их медиа группой
	if len(photoContent) != 0 {
		_, err = tgBot.Send(tgbotapi.NewMediaGroup(telegramChatId, photoContent))

		if err != nil {
			return err
		}

		return nil
	}

	// Если у ВК поста нет фотографий - просто шлём сообщение текстом (нет обработки музыки, видео отправляется в виде ссылки на vk)
	_, err = tgBot.Send(tgbotapi.NewMessage(telegramChatId, preparedMsg))

	if err != nil {
		return err
	}

	return nil
}

var (
	vkGroupUrl    *string
	tgToken       *string
	tgChat        *int64
	updateTimeout *int
)

func init() {
	vkGroupUrl = flag.String("vk-url", "", "Ссылка на VK паблик")
	tgToken = flag.String("tg-token", "", "Токен от телеграмм бота")
	tgChat = flag.Int64("tg-chat", 0, "Телеграм чат ID")
	updateTimeout = flag.Int("update-timeout", 10, "Частота обновления данных в боте (сек)")
}

func validateCliArgs() {
	if *vkGroupUrl == "" {
		panic("Не задана ссылка на vk паблик")
	}

	if *tgToken == "" {
		panic("Не задан телеграмм токен")
	}

	if *tgChat == 0 {
		panic("Не задан телеграмм чат")
	}
}

func main() {
	// Получаем аргументы командной строки
	flag.Parse()
	validateCliArgs()

	// Хардкод redis сервиса
	redisConn := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisConn.Ping().Result()

	if err != nil {
		fmt.Println("Не удалось подключиться к Redis СУБД, причина: ", err)
		os.Exit(2)
	}

	defer redisConn.Close()

	tgBot, _ := tgbotapi.NewBotAPI(*tgToken)
	httpClient := &http.Client{}

	for true {
		file, error := httpClient.Get(*vkGroupUrl)

		if error != nil {
			fmt.Println("Не удалось получить данные из Vk паблика, ошибка: ", error)
			os.Exit(1)
		}

		vkPosts := VkParser.GetVkPosts(file.Body)

		for _, vkPost := range vkPosts {
			// Если в мы уже публиковали данный пост - пропускаем итерацию
			if redisConn.Exists([]string{vkPost.GetId()}...).Val() != 0 {
				continue
			}

			// публикуем вк пост в телеграмм
			err := publishVkPostToTelegram(*tgChat, vkPost, tgBot)

			// Если в ходе публикация поста произошла ошибка - пишем её в stdout и пропускаем итерацию цикла
			if err != nil {
				fmt.Println(err)
				continue
			}

			// записываем в Redis метку что пост выл выгружен в Telegram
			redisConn.Set(vkPost.GetId(), 1, 0)
		}

		time.Sleep(time.Second * time.Duration(*updateTimeout))
	}
}

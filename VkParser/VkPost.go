package VkParser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
)

type VkPost struct {
	id           string
	text         string
	videosLinks  []string
	pictureLinks []string
}

func (vkPost *VkPost) GetId() string {
	return vkPost.id
}

func (vkPost *VkPost) GetText() string {
	return vkPost.text
}

func (vkPost *VkPost) GetVideoLinks() []string {
	return vkPost.videosLinks
}

func (vkPost *VkPost) GetPictureLinks() []string {
	return vkPost.pictureLinks
}

func newVkPost(id string, text string) VkPost {
	return VkPost{
		id:   id,
		text: text,
	}
}

func (vkPost *VkPost) addVideoLink(videoLink string) {
	vkPost.videosLinks = append(vkPost.videosLinks, videoLink)
}

func (vkPost *VkPost) addPictureLink(pictureLink string) {
	vkPost.pictureLinks = append(vkPost.pictureLinks, pictureLink)
}

func GetVkPosts(postsReader io.Reader) []VkPost {
	vkDomain := "https://vk.com"
	doc, err := goquery.NewDocumentFromReader(postsReader)
	if err != nil {
		log.Fatal(err)
	}

	VkPosts := make([]VkPost, 0)

	doc.Find(".wall_item").Each(func(i int, s *goquery.Selection) {

		postId, existId := s.Find(".post__anchor.anchor").Attr("name")

		if !existId {
			fmt.Println("У поста не найден идентификатор в верстке")
			return
		}

		postText := s.Find(".pi_text").Text()
		vkPost := newVkPost(postId, postText)

		s.Find(".thumb_map.thumb_map_wide.thumb_map_l.al_photo > a[aria-label*=фотография]").Each(func(i int, selection *goquery.Selection) {
			pictureLink, existPictureAttachment := selection.Attr("href")

			if existPictureAttachment {
				vkPost.addPictureLink(vkDomain + pictureLink)
			}
		})

		s.Find(".thumb_map.thumb_map_wide.thumb_map_l > a[aria-label*=Видео]").Each(func(i int, selection *goquery.Selection) {
			videoLink, existVideoAttachment := selection.Attr("href")

			if existVideoAttachment {
				vkPost.addVideoLink(vkDomain + videoLink)
			}
		})

		VkPosts = append(VkPosts, vkPost)
	})

	return VkPosts
}

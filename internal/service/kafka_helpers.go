package service

import (
	"context"
	"fmt"

	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/constant"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"
)

// publishEvent 发布 Kafka 事件的通用方法
func publishEvent(topic, key string, event pkgKafka.Event) {
	kw := global.GetKafka()
	if kw == nil {
		return
	}
	if err := kw.Publish(context.Background(), key, event); err != nil {
		fmt.Printf("[kafka] publish to %s failed: %v\n", topic, err)
	}
}

// publishPostEvent 发布帖子相关事件
func publishPostEvent(action string, userId, postId uint) {
	publishEvent(pkgKafka.TopicPostEvent, fmt.Sprintf("post:%d", postId), pkgKafka.Event{
		Action:     action,
		UserID:     userId,
		TargetID:   postId,
		TargetType: "post",
	})
}

// publishCommentEvent 发布评论相关事件
func publishCommentEvent(action string, userId, commentId uint, targetType constant.TargetType, targetId uint) {
	publishEvent(pkgKafka.TopicPostEvent, fmt.Sprintf("comment:%d", commentId), pkgKafka.Event{
		Action:     action,
		UserID:     userId,
		TargetID:   commentId,
		TargetType: string(targetType),
	})
}

// publishNotification 发布通知事件
func publishNotification(action string, userId, targetId uint, targetType string, receiverId uint, content string) {
	publishEvent(pkgKafka.TopicNotification, fmt.Sprintf("notif:%d", receiverId), pkgKafka.Event{
		Action:     action,
		UserID:     userId,
		TargetID:   targetId,
		TargetType: targetType,
		Extra: pkgKafka.NotificationEvent{
			ReceiverID: receiverId,
			Content:    content,
		},
	})
}

// publishPostNotification 发布帖子相关通知（点赞/收藏等）
func publishPostNotification(action string, userId, postId uint, content string) {
	post, err := getPostById(postId)
	if err != nil {
		return
	}
	if post.UserId == userId {
		return // 不给自己发通知
	}
	publishNotification(action, userId, postId, "post", post.UserId, content)
}

// publishCommentNotification 发布评论通知
func publishCommentNotification(userId, commentId uint, targetType constant.TargetType, targetId uint, content string) {
	var ownerId uint
	switch targetType {
	case constant.TargetPost:
		post, err := getPostById(targetId)
		if err != nil {
			return
		}
		ownerId = post.UserId
	case constant.TargetVideo:
		// TODO: 获取视频作者
		return
	default:
		return
	}
	if ownerId == userId {
		return // 不给自己发通知
	}
	publishNotification(pkgKafka.ActionComment, userId, targetId, string(targetType), ownerId, content)
}

// publishFollowNotification 发布关注通知
func publishFollowNotification(followerId, followingId uint) {
	publishNotification(pkgKafka.ActionFollow, followerId, followingId, "user", followingId,
		fmt.Sprintf("用户 %d 关注了你", followerId))
}

// publishCacheInvalidation 发布缓存失效事件
func publishCacheInvalidation(targetType string, targetId uint) {
	publishEvent(pkgKafka.TopicUserEvent, fmt.Sprintf("cache:%s:%d", targetType, targetId), pkgKafka.Event{
		Action:     "invalidate",
		TargetID:   targetId,
		TargetType: targetType,
	})
}

// getPostById 获取帖子（内部辅助函数）
func getPostById(id uint) (*entity.Post, error) {
	var post entity.Post
	err := global.GetDB().Where("id = ?", id).First(&post).Error
	return &post, err
}

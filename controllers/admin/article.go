package admin

import (
	"github.com/astaxie/beego/orm"
	"github.com/lisijie/goblog/models"
	"strconv"
	"strings"
	"time"
)

type ArticleController struct {
	baseController
}

//管理
func (this *ArticleController) List() {
	page, _ := strconv.ParseInt(this.Ctx.Input.Param(":page"), 10, 0)
	if page < 1 {
		page = 1
	}
	pagesize := int64(10)
	offset := (page - 1) * pagesize

	var list []*models.Post
	var post models.Post
	o := orm.NewOrm()
	count, _ := o.QueryTable(&post).Count()
	if count > 0 {
		o.QueryTable(&post).Limit(pagesize, offset).All(&list)
	}

	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pagebar"] = models.Pager(page, count, pagesize, "/admin/article/list")
	this.display()
}

//添加
func (this *ArticleController) Add() {
	this.display()
}

//编辑
func (this *ArticleController) Edit() {
	id, _ := this.GetInt("id")
	post, _ := models.GetPost(id)
	this.Data["post"] = post
	this.display()
}

//保存
func (this *ArticleController) Save() {
	o := orm.NewOrm()

	id, _ := this.GetInt("id")
	title := this.GetString("title")
	content := this.GetString("content")
	tags := this.GetString("tags")

	addtags := make([]string, 0)
	//标签过滤
	if tags != "" {
		tagarr := strings.Split(tags, ",")
		for _, v := range tagarr {
			if tag := strings.TrimSpace(v); tag != "" {
				exists := false
				for _, vv := range addtags {
					if vv == tag {
						exists = true
						break
					}
				}
				if !exists {
					addtags = append(addtags, tag)
				}
			}
		}
	}

	post := new(models.Post)
	if id < 1 {
		post.Userid = this.userid
		post.Author = this.username
		post.Posttime = time.Now()
		post.Title = title
		post.Content = content
		post.Id, _ = orm.NewOrm().Insert(post)
	} else {
		post.Id = id
		if o.Read(post) != nil {
			goto RD
		}
		post.Title = title
		post.Content = content
		if post.Tags != "" {
			oldtags := strings.Split(post.Tags, ",")
			//标签统计-1
			o.QueryTable(&models.Tag{}).Filter("name__in", oldtags).Update(orm.Params{"count": orm.ColValue(orm.Col_Minus, 1)})
			//删掉tag_post表的记录
			o.QueryTable(&models.TagPost{}).Filter("postid", post.Id).Delete()
		}
	}

	if len(addtags) > 0 {
		for _, v := range addtags {
			tag := new(models.Tag)
			tag.Name = v
			if o.Read(tag, "Name") == orm.ErrNoRows {
				tag.Count = 1
				tag.Id, _ = o.Insert(tag)
			} else {
				tag.Count += 1
				o.Update(tag)
			}

			tp := new(models.TagPost)
			tp.Tagid = tag.Id
			tp.Postid = post.Id
			o.Insert(tp)
		}
		post.Tags = strings.Join(addtags, ",")
	}
	o.Update(post)

RD:
	this.Redirect("/admin/article/list", 302)
}

//删除
func (this *ArticleController) Delete() {
	id, _ := this.GetInt("id")

	post := new(models.Post)
	post.Id = id
	o := orm.NewOrm()
	if o.Read(post) == nil {
		if post.Tags != "" {
			oldtags := strings.Split(post.Tags, ",")
			//标签统计-1
			o.QueryTable(&models.Tag{}).Filter("name__in", oldtags).Update(orm.Params{"count": orm.ColValue(orm.Col_Minus, 1)})
			//删掉tag_post表的记录
			o.QueryTable(&models.TagPost{}).Filter("postid", post.Id).Delete()
		}
		o.Delete(post)
	}
	this.Redirect("/admin/article/list", 302)
}

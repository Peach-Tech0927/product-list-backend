package main

import "github.com/gin-gonic/gin"
import "net/http"
import "database/sql"//GORMを作る際に必要無くなる？
import _ "github.com/lib/pq"//GORMを作る際に必要無くなる？
import "encoding/base64"
import "log"
import "io"
import "strconv"

type Content struct {
    ID          int
    Creator     string
    Title       string
    Description string
    Image       string
}

func main() {
    db, err := sql.Open("postgres", "postgres://sugatario:@localhost:5432/crud?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    r := gin.Default()
    r.LoadHTMLGlob("templates/*")

    r.StaticFile("/favicon.ico", "templates")
    r.MaxMultipartMemory = 8 << 20;

    r.GET("/", func(c *gin.Context) {
        var contents []Content
        rows, err := db.Query("select id, creator, title, description from contents")
        if err != nil {
            log.Fatal(err)
        }
        defer rows.Close()

        for rows.Next() {
            var content Content
            err = rows.Scan(&content.ID, &content.Creator, &content.Title, &content.Description)
            if err != nil {
                log.Fatal(err)
            }
            contents = append(contents, content)
        }
        err = rows.Err()
        if err != nil {
            log.Fatal(err)
        }
        // TODO: creator/title/descriptionが表示されているか確認する
        c.HTML(http.StatusOK, "index.tmpl", gin.H {
            "contents": contents,
        })
    })

    r.GET("/image/:id", func(c *gin.Context) {
        id := c.Param("id")

        var content Content

        stmt, err := db.Prepare("select image from contents where id = $1")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        err = stmt.QueryRow(id).Scan(&content.Image)
        if err != nil {
            log.Fatal(err)
        }

        dec, err :=  base64.StdEncoding.DecodeString(content.Image)
        if err != nil {
            log.Fatal(err)
        }

        c.Data(http.StatusOK, "image/jpeg", dec)
    })

    r.GET("/:id", func(c *gin.Context) {
        id, _ := strconv.Atoi(c.Param("id"))

        stmt, err := db.Prepare("select creator, title, description from contents where id = $1")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        var content Content

        err = stmt.QueryRow(id).Scan(&content.Creator, &content.Title, &content.Description)
        if err != nil {
            log.Fatal(err)
        }
        content.ID = id
        
        c.HTML(http.StatusOK, "detail.tmpl", gin.H {
            "content": content,
        })
    })
    
    r.POST("/", func(c *gin.Context) {
        var content Content

        content.Creator = c.PostForm("creator")
        content.Title = c.PostForm("title")
        content.Description = c.PostForm("description")
        
        image, err := c.FormFile("image")
        if err != nil {
            log.Fatal(err)
        }

        f, err := image.Open()
        if err != nil {
            log.Fatal(err)
        }
        defer f.Close()

        image_data, err := io.ReadAll(f)
        if err != nil {
            log.Fatal(err)
        }

        content.Image = base64.StdEncoding.EncodeToString(image_data)

        tx, err := db.Begin()
        if err != nil {
            log.Fatal(err)
        }
        defer tx.Rollback()

        stmt, err := tx.Prepare("insert into contents(creator, title, description, image) values($1, $2, $3, $4)")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        _, err = stmt.Exec(content.Creator, content.Title, content.Description, content.Image)
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Commit()
        if err != nil {
            log. Fatal(err)
        }

        c.HTML(http.StatusOK, "detail.tmpl", gin.H {
            "creator": content.Creator,
            "title": content.Title,
            "description": content.Description,
        })
    })

    r.POST("/:id/update", func(c *gin.Context) {
        id, _ := strconv.Atoi(c.Param("id"))
        var content Content

        content.ID = id
        content.Creator = c.PostForm("creator")
        content.Title = c.PostForm("title")
        content.Description = c.PostForm("description")
        
        image, err := c.FormFile("image")
        if err != nil {
            log.Fatal(err)
       }

        f, err := image.Open()
        if err != nil {
            log.Fatal(err)
        }
        defer f.Close()

        image_data, err := io.ReadAll(f)
        if err != nil {
            log.Fatal(err)
        }

        content.Image = base64.StdEncoding.EncodeToString(image_data)

        tx, err := db.Begin()
        if err != nil {
            log.Fatal(err)
        }
        defer tx.Rollback()

        stmt, err := tx.Prepare("update contents set creator = $1, title =$2, description = $3, image =$4 where id = $5")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        _, err = stmt.Exec(content.Creator, content.Title, content.Description, content.Image, id)
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Commit()
        if err != nil {
            log.Fatal(err)
        }

        c.HTML(http.StatusOK, "updated.tmpl", gin.H {
            "content": content,
        })
    })

    r.POST("/:id/delete", func(c *gin.Context) {
        id := c.Param("id")

        tx, err := db.Begin()
        if err != nil {
            log.Fatal(err)
        }
        defer tx.Rollback()

        stmt, err := tx.Prepare("DELETE FROM contents WHERE id = $1")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        _, err = stmt.Exec(id)
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Commit()
        if err != nil {
            log.Fatal(err)
        }

        c.HTML(http.StatusOK, "deleted.tmpl", gin.H {
            "id": id,
        })
    })

    r.Run()
}


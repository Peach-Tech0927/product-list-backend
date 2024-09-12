package main

import "github.com/gin-gonic/gin"
import "net/http"
import "database/sql"
import _ "github.com/lib/pq"
import "encoding/base64"
import "log"
import "io"

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
        contents := make(map[int]string)
        rows, err := db.Query("select id, creator, title, description from contents")
        if err != nil {
            log.Fatal(err)
        }
        defer rows.Close()

        for rows.Next() {
            var id int
            var creator string
            var title string
            var description string
            err = rows.Scan(&id, &creator, &title, &description)
            if err != nil {
                log.Fatal(err)
            }
            contents[id] = creator
            contents[id] = title
            contents[id] = description
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

        stmt, err := db.Prepare("select image from contents where id = $1")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        var image string

        err = stmt.QueryRow(id).Scan(&image)
        if err != nil {
            log.Fatal(err)
        }

        dec, err :=  base64.StdEncoding.DecodeString(image)
        if err != nil {
            log.Fatal(err)
        }

        c.Data(http.StatusOK, "image/jpeg", dec)
    })

    r.GET("/:id", func(c *gin.Context) {
        id := c.Param("id")

        stmt, err := db.Prepare("select creator, title, description from contents where id = $1")
        if err != nil {
            log.Fatal(err)
        }
        defer stmt.Close()

        var creator string
        var title string
        var description string

        err = stmt.QueryRow(id).Scan(&creator, &title, &description)
        if err != nil {
            log.Fatal(err)
        }
        
        c.HTML(http.StatusOK, "detail.tmpl", gin.H {
            "creator": creator,
            "title": title,
            "description": description,
        })
    })
    
    r.POST("/", func(c *gin.Context) {
        creator := c.PostForm("creator")
        title := c.PostForm("title")
        description := c.PostForm("description")
        
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

        encoded_image_data := base64.StdEncoding.EncodeToString(image_data)

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

        _, err = stmt.Exec(creator, title, description, encoded_image_data)
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Commit()
        if err != nil {
            log. Fatal(err)
        }

        c.HTML(http.StatusOK, "detail.tmpl", gin.H {
            "creator": creator,
            "title": title,
            "description": description,
        })
    })

    r.POST("/:id/update", func(c *gin.Context) {
        id := c.Param("id")
        creator := c.PostForm("creator")
        title := c.PostForm("title")
        description := c.PostForm("description")
        
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

        encoded_image_data := base64.StdEncoding.EncodeToString(image_data)

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

        _, err = stmt.Exec(creator, title, description, encoded_image_data, id)
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Commit()
        if err != nil {
            log.Fatal(err)
        }

        c.HTML(http.StatusOK, "updated.tmpl", gin.H {
            "creator": creator,
            "title": title,
            "description": description,
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


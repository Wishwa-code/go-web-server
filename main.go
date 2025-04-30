// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hello is a simple hello, world demonstration web server.
//
// It serves version information on /version and answers
// any other request like /name by saying "Hello, name!".
//
// See golang.org/x/example/outyet for a more sophisticated server.

//*-----------------------------------------------------------------*//
//*---------in case of emergency uncomment all the line marked green *//
//*-----------------------------------------------------------------*//

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func Hello(name string) string {
	// Return a greeting that embeds the name in a message.
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}

type Brand struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Logourl          string     `json:"logourl"`
	CountryOfOrigin  string     `json:"country_of_origin"`
	SocialMediaLinks string     `json:"social_media_links"`
	ContactEmail     string     `json:"contact_email"`
	PhoneNumber      string     `json:"phone_number"`
	BannerUrl        string     `json:"banner_url"`
	Website          string     `json:"website"`
	CreatedAt        *time.Time `json:"created_at"`
}

type Supplier struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	LogoURL          string     `json:"logourl"`
	Website          string     `json:"website"`
	CountryOfOrigin  string     `json:"coutry_of_origin"`
	SocialMediaLinks string     `json:"social_media_links"`
	ContactEmail     string     `json:"contact_email"`
	PhoneNumber      string     `json:"phone_number"`
	BannerURL        string     `json:"banner_url"`
	LocatedCity      string     `json:"city"`
	LocatedCountry   string     `json:"country"`
	BankDetails      string     `json:"bank_details"`
	Status           string     `json:"status"`
	ExtraData        string     `json:"extra_data"`
	CreatedAt        *time.Time `json:"created_at"`
}

type Product struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	TagOne         string     `json:"tag_one"`
	TagTwo         string     `json:"tag_two"`
	ImageURL       string     `json:"imageurl"`
	Department     string     `json:"department"`
	MainCategory   string     `json:"main_catogory"`
	SubCategory    string     `json:"sub_catogory"`
	CreatedAt      *time.Time `json:"created_at"`
	LastModifiedAt *time.Time `json:"last_modified_at"`
}

type Variance struct {
	ID                  int        `json:"id"` // pointer to differentiate between null and 0
	ProductName         string     `json:"productName"`
	ProductID           string     `json:"product_id"`
	Barcode             string     `json:"barcode"`
	DisplayTitle        string     `json:"displayTitle"`
	VarianceDescription string     `json:"about_this_variance"`
	ImageUrl            string     `json:"imageurl"`
	VarianceTitle       string     `json:"variance"`
	Brand               string     `json:"brand"`
	Supplier            string     `json:"supplier"`
	OriginalPrice       float64    `json:"original_price"`         // Changed to float64 for NUMERIC
	RetailPrice         float64    `json:"retail_price"`           // Changed to float64 for NUMERIC
	WholesalePrice      float64    `json:"wholesale_price"`        // Changed to float64 for NUMERIC
	Quantity            float64    `json:"quantity"`               // new field
	UnitMeasure         string     `json:"unit_measure"`           // DOUBLE PRECISION
	LeastSubUnitMeasure float64    `json:"least_sub_unit_measure"` // text
	CreatedAt           *time.Time `json:"created_at"`
	LastModifiedAt      *time.Time `json:"last_modified_at"`
}

var memoryDb = make(map[string]string)
var postgresDb *sql.DB

//! ============================================================================ //
//? ========================== 🫰 THE MAIN 🫰 ================================= //
//! ============================================================================ //

func main() {

	message := Hello("Models imported 🐳")
	fmt.Println(message)

	connStr := "postgresql://redrose_owner:npg_bxEKF6r9hvJu@ep-empty-dew-a13g9uj0-pooler.ap-southeast-1.aws.neon.tech/redrose?sslmode=require"
	var err error
	postgresDb, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer postgresDb.Close()
	var version string
	if err := postgresDb.QueryRow("select version()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Printf("Connected to db version=%s\n", version)

	r := setupRouter()

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")

}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := memoryDb[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			memoryDb[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	// Add getAllProducts endpoint
	r.POST("/products/insert", insertProduct)

	r.GET("/products", getAllProducts)

	r.GET("/products/search", searchProducts)

	r.GET("/products/get-product/:id", getProductByID)

	r.PUT("/products/update", updateProduct)

	r.GET("/products/last-product", getLastProduct)

	r.POST("/variance/upsert", insertOrUpdateVariance)

	r.GET("/variance/last", getLastVariance)

	r.GET("/variance/by-product/:id", getVariancesByProductId)

	r.POST("/supplier/upsert", insertOrUpdateSupplier)

	r.GET("/supplier/getAll", getSupplierFilters)

	r.POST("brand/upsert", insertOrUpdateBrand)

	r.GET("/brand/getAll", getBrandFilters)

	return r
}

//? ========================================================================= //
//! ================== 📦 PRODUCT RELATED API HANDLERS 📦 ================== //
//? ========================================================================= //

func insertProduct(c *gin.Context) {
	// Define struct to bind incoming JSON

	var product Product
	// Bind JSON input
	if err := c.ShouldBindJSON(&product); err != nil {
		log.Println("Found error when parsing json", err)

		c.JSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_JSON",
					"message": "Invalid JSON input",
					"details": err.Error(),
				},
			})

		return
	}

	// Generate ID if not provided
	if product.ID == "" {
		product.ID = gofakeit.UUID()
	}

	now := time.Now()
	product.CreatedAt = &now
	product.LastModifiedAt = &now

	// Insert into database
	_, err := postgresDb.Exec(`
		INSERT INTO products (id, title, description, tag_one, tag_two, imageurl, department, main_catogory, sub_catogory, created_at, last_modified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, product.ID, product.Title, product.Description, product.TagOne, product.TagTwo, product.ImageURL,
		product.Department, product.MainCategory, product.SubCategory, product.CreatedAt, product.LastModifiedAt)

	if err != nil {
		log.Println("Found error while performing db query", err)

		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	// Respond with inserted product
	c.JSON(http.StatusOK, gin.H{
		"status":  "product inserted",
		"product": product,
	})
}

func getLastProduct(c *gin.Context) {
	// Query to find the last product added
	query := `
		SELECT 
			id, 
			COALESCE(title, '') AS title, 
			COALESCE(description, '') AS description, 
			COALESCE(tag_one, 'N/A') AS tag_one, 
			COALESCE(tag_two, 'N/A') AS tag_two,  
			COALESCE(imageurl, '') AS imageurl, 
			COALESCE(department, 'mainBuilding') AS department, 
			COALESCE(main_catogory, 'sand') AS main_catogory, 
			COALESCE(sub_catogory, 'N/A') AS sub_catogory,
			created_at,
			last_modified_at 
		FROM products
		ORDER BY last_modified_at DESC
		LIMIT 1
	`

	// Execute the query
	row := postgresDb.QueryRow(query)

	// Define a struct to hold the product
	var product Product

	// Scan the result into the product struct
	err := row.Scan(
		&product.ID, &product.Title, &product.Description, &product.TagOne, &product.TagTwo,
		&product.ImageURL, &product.Department,
		&product.MainCategory, &product.SubCategory, &product.CreatedAt, &product.LastModifiedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest,
				gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_ROWS",
						"message": "No rows found",
						"details": err.Error(),
					},
				})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	// Return the result as JSON
	c.JSON(http.StatusOK, gin.H{"product": product})
}

func getAllProducts(c *gin.Context) {
	// Use COALESCE to replace NULL values with a default enum value for enum fields
	rows, err := postgresDb.Query(`
		SELECT 
			id, 
			COALESCE(title, '') AS title, 
			COALESCE(description, '') AS description, 
			COALESCE(tag_one, 'N/A') AS tag_one, 
			COALESCE(tag_two, 'N/A') AS tag_two,  
			COALESCE(imageurl, '') AS imageurl, 
			COALESCE(department, 'mainBuilding') AS department, 
			COALESCE(main_catogory, 'sand') AS main_catogory, 
			COALESCE(sub_catogory, 'N/A') AS sub_catogory,
			created_at,
			last_modified_at 
		FROM product
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to fetch products",
					"details": err.Error(),
				},
			})
		return
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Title, &product.Description, &product.TagOne, &product.TagTwo, &product.ImageURL, &product.Department, &product.MainCategory, &product.SubCategory, &product.CreatedAt, &product.LastModifiedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to process data returned from database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func searchProducts(c *gin.Context) {
	title := c.Query("title")
	brand := c.Query("brand")
	department := c.Query("department")
	mainCatogory := c.Query("main_catogory")
	subCatogory := c.Query("sub_catogory")
	sort := c.DefaultQuery("sort", "title")
	order := c.DefaultQuery("order", "asc")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pagesize", "10")
	lookInDescription := c.DefaultQuery("lookinDescription", "false")

	//* log.Println("This is a title", title)
	// Parse pagination params
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 {
		pageSizeNum = 10
	}
	offset := (pageNum - 1) * pageSizeNum

	query := `
			SELECT 
			id, 
			COALESCE(title, '') AS title, 
			COALESCE(description, '') AS description, 
			COALESCE(tag_one, 'N/A') AS tag_one, 
			COALESCE(tag_two, 'N/A') AS tag_two,  
			COALESCE(imageurl, '') AS imageurl, 
			COALESCE(department, 'mainBuilding') AS department, 
			COALESCE(main_catogory, 'sand') AS main_catogory, 
			COALESCE(sub_catogory, 'N/A') AS sub_catogory,
			created_at,
			last_modified_at 
		FROM products
		WHERE 1=1
	`
	args := []interface{}{}
	argID := 1

	if title != "" {
		if strings.ToLower(lookInDescription) == "true" {
			query += fmt.Sprintf(" AND (LOWER(title) LIKE LOWER($%d) OR LOWER(description) LIKE LOWER($%d))", argID, argID+1)
			args = append(args, fmt.Sprintf("%%%s%%", title), fmt.Sprintf("%%%s%%", title))
			argID += 2
		} else {
			query += fmt.Sprintf(" AND LOWER(title) LIKE LOWER($%d)", argID)
			args = append(args, fmt.Sprintf("%%%s%%", title))
			argID++
		}
	}
	if brand != "" {
		query += fmt.Sprintf(" AND LOWER(brand) = LOWER($%d)", argID)
		args = append(args, brand)
		argID++
	}
	if department != "" {
		departments := strings.Split(department, ",")
		if len(departments) > 0 {
			placeholders := []string{}
			for _, dept := range departments {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argID))
				args = append(args, strings.TrimSpace(dept))
				argID++
			}
			query += fmt.Sprintf(" AND department IN (%s)", strings.Join(placeholders, ", "))
		} else {
			// Handle empty department case
			query += " AND department IN ('')"
		}
	}
	if mainCatogory != "" {
		mainCatogory := strings.Split(mainCatogory, ",")
		if len(mainCatogory) > 0 {
			placeholders := []string{}
			for _, dept := range mainCatogory {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argID))
				args = append(args, strings.TrimSpace(dept))
				argID++
			}
			query += fmt.Sprintf(" AND main_catogory IN (%s)", strings.Join(placeholders, ", "))
		} else {
			query += " AND main_catogory IN ('')"
		}
	}

	if subCatogory != "" {
		subCatogory := strings.Split(subCatogory, ",")
		if len(subCatogory) > 0 {
			placeholders := []string{}
			for _, dept := range subCatogory {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argID))
				args = append(args, strings.TrimSpace(dept))
				argID++
			}
			query += fmt.Sprintf(" AND sub_catogory IN (%s)", strings.Join(placeholders, ", "))
		} else {
			query += " AND sub_catogory IN ('')"
		}
	}

	// Sorting
	query += fmt.Sprintf(" ORDER BY %s %s", sort, order)

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, pageSizeNum, offset)

	query = fmt.Sprintf("-- dynamic | arg count: %d\n%s", len(args), query)

	rows, err := postgresDb.Query(query, args...)

	///* log.Println("This is a rows returned", rows)

	if err != nil {
		if err == sql.ErrNoRows {
			// log.Println("Found error --> product by id | state:just called db", err)
			c.JSON(http.StatusBadRequest,
				gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_ROWS",
						"message": "No rows found",
						"details": err.Error(),
					},
				})
			return
		}
		log.Printf("SQL: %s | ARGS: %#v", query, args)

		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.TagOne, &p.TagTwo,
			&p.ImageURL, &p.Department, &p.MainCategory, &p.SubCategory, &p.CreatedAt, &p.LastModifiedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	//* log.Println("This is a title", products)

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func getProductByID(c *gin.Context) {
	id := c.Param("id")

	query := `
		SELECT 
			id, 
			COALESCE(title, '') AS title, 
			COALESCE(description, '') AS description, 
			COALESCE(tag_one, 'N/A') AS tag_one, 
			COALESCE(tag_two, 'N/A') AS tag_two,  
			COALESCE(imageurl, '') AS imageurl, 
			COALESCE(department, 'mainBuilding') AS department, 
			COALESCE(main_catogory, 'sand') AS main_catogory, 
			COALESCE(sub_catogory, 'N/A') AS sub_catogory,
			created_at,
			last_modified_at 
		FROM products 
		WHERE id = $1
	`

	var product Product

	log.Printf("Executing SQL: %s with id = %s", query, id)

	err := postgresDb.QueryRow(query, id).Scan(
		&product.ID, &product.Title, &product.Description,
		&product.TagOne, &product.TagTwo, &product.ImageURL,
		&product.Department, &product.MainCategory, &product.SubCategory, &product.CreatedAt, &product.LastModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Found error --> product by id | state:just called db", err)
			c.JSON(http.StatusBadRequest,
				gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_ROWS",
						"message": "No rows found",
						"details": err.Error(),
					},
				})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	log.Printf("Calling getProductByID with param id = %s", id)

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func updateProduct(c *gin.Context) {
	var product Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_JSON",
					"message": "Invalid JSON input",
					"details": err.Error(),
				},
			})
		return
	}

	now := time.Now()
	product.LastModifiedAt = &now

	query := `
		UPDATE products
		SET 
			title = $1,
			description = $2,
			tag_one = $3,
			tag_two = $4,
			imageurl = $5,
			department = $6,
			main_catogory = $7,
			sub_catogory = $8,
			last_modified_at = $9
		WHERE id = $10
	`

	result, err := postgresDb.Exec(query,
		product.Title, product.Description, product.TagOne,
		product.TagTwo, product.ImageURL, product.Department,
		product.MainCategory, product.SubCategory, product.LastModifiedAt, product.ID,
	)

	if err != nil {

		log.Println("Found errr when updating product(category) | state:just called db", err, product.ID)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to update product into database",
					"details": err.Error(),
				},
			})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "No rows were affected by update",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "product updated",
		"product": product,
	})
}

//! ============================================================================ //
//? ================= ✨ PRODUCT VARIANCE RELATED API HANDLERS ✨ =============== //
//! ============================================================================ //

func insertOrUpdateVariance(c *gin.Context) {
	var v Variance
	if err := c.ShouldBindJSON(&v); err != nil {
		log.Println("📢 upserting variances to json parsing got error", err)

		c.JSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_JSON",
					"message": "Invalid JSON input",
					"details": err.Error(),
				},
			})

		return
	}

	now := time.Now()
	v.CreatedAt = &now
	v.LastModifiedAt = &now

	query := `
		INSERT INTO products_variances (
			images, original_price, retail_price, wholesale_price,
			about_this_variance, variance_display_title, product, variance, brand_name,
			product_id, supplier, quantity, unit_measure, least_sub_unit_measure, barcode, created_at, last_modified_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (product, variance, brand_name)
		DO UPDATE SET 
			original_price = EXCLUDED.original_price,
			retail_price = EXCLUDED.retail_price,
			wholesale_price = EXCLUDED.wholesale_price,
			about_this_variance = EXCLUDED.about_this_variance,
			variance_display_title = EXCLUDED.variance_display_title,
			supplier = EXCLUDED.supplier,
			quantity = EXCLUDED.quantity,
			unit_measure = EXCLUDED.unit_measure,
			least_sub_unit_measure = EXCLUDED.least_sub_unit_measure,
			images = EXCLUDED.images,
			barcode = EXCLUDED.barcode,
			last_modified_at = EXCLUDED.last_modified_at
		RETURNING id, images, original_price, retail_price, wholesale_price,
		          about_this_variance, variance_display_title, product, variance, brand_name,
		          product_id, supplier, quantity, unit_measure, least_sub_unit_measure, barcode, created_at, last_modified_at
	`

	// JSON encode the image URL
	imageJson := fmt.Sprintf(`["%s"]`, v.ImageUrl)

	var result Variance
	var images string

	err := postgresDb.QueryRow(
		query,
		imageJson, v.OriginalPrice, v.RetailPrice, v.WholesalePrice,
		v.VarianceDescription, v.DisplayTitle, v.ProductName, v.VarianceTitle, v.Brand,
		v.ProductID, v.Supplier, v.Quantity, v.UnitMeasure, v.LeastSubUnitMeasure, v.Barcode, v.CreatedAt, v.LastModifiedAt,
	).Scan(
		&result.ID, &images, &result.OriginalPrice, &result.RetailPrice, &result.WholesalePrice,
		&result.VarianceDescription, &result.DisplayTitle, &result.ProductName, &result.VarianceTitle,
		&result.Brand, &result.ProductID, &result.Supplier, &result.Quantity, &result.UnitMeasure,
		&result.LeastSubUnitMeasure, &result.Barcode, &result.CreatedAt, &result.LastModifiedAt,
	)

	if err != nil {
		log.Println("📢 upserting variances to db got error", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "variance upserted",
		"product": result,
	})
}

func getLastVariance(c *gin.Context) {
	query := `
		SELECT 
			id,
			COALESCE(product, '') AS product,
			product_id,
			COALESCE(variance_display_title, '') AS variance_display_title,
			COALESCE(about_this_variance, '') AS about_this_variance,
			COALESCE(images->>0, '') AS imageurl, -- get first image URL from JSON array
			COALESCE(variance, '') AS variance,
			COALESCE(brand_name, '') AS brand,
			COALESCE(supplier, '') AS supplier,
			COALESCE(original_price, 0) AS original_price,
			COALESCE(retail_price, 0) AS retail_price,
			COALESCE(wholesale_price, 0) AS wholesale_price, 	
			COALESCE(quantity, 0) AS quantity,
			COALESCE(unit_measure, '') AS unit_measure,
			COALESCE(least_sub_unit_measure, 0) AS least_sub_unit_measure,
			COALESCE(barcode, '') AS barcode,
			created_at,
			last_modified_at
		FROM products_variances
		ORDER BY last_modified_at DESC
		LIMIT 1
	`

	row := postgresDb.QueryRow(query)

	var v Variance
	err := row.Scan(
		&v.ID, &v.ProductName, &v.ProductID, &v.DisplayTitle,
		&v.VarianceDescription, &v.ImageUrl, &v.VarianceTitle,
		&v.Brand, &v.Supplier, &v.OriginalPrice,
		&v.RetailPrice, &v.WholesalePrice, &v.Quantity,
		&v.UnitMeasure,
		&v.LeastSubUnitMeasure, &v.Barcode, &v.CreatedAt, &v.LastModifiedAt,
	)

	if err != nil {
		log.Println("📢 Error from fetch last variance query: ", err)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest,
				gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_ROWS",
						"message": "No rows found",
						"details": err.Error(),
					},
				})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variance": v})
}

func getVariancesByProductId(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	query := `
		SELECT 
			id,
			COALESCE(product, '') AS product,
			product_id,
			COALESCE(variance_display_title, '') AS variance_display_title,
			COALESCE(about_this_variance, '') AS about_this_variance,
			COALESCE(images->>0, '') AS imageurl,
			COALESCE(variance, '') AS variance,
			COALESCE(brand_name, '') AS brand,
			COALESCE(supplier, '') AS supplier,
			COALESCE(original_price, 0) AS original_price,
			COALESCE(retail_price, 0) AS retail_price,
			COALESCE(wholesale_price, 0) AS wholesale_price,
			COALESCE(quantity, 0) AS quantity,
			COALESCE(unit_measure, '') AS unit_measure,
			COALESCE(least_sub_unit_measure, 0) AS least_sub_unit_measure,
			COALESCE(barcode, '') AS barcode,
			created_at,
			last_modified_at 
		FROM products_variances
		WHERE product_id = $1
		ORDER BY id DESC
	`
	//* log.Printf("Running getVariancesByProductId query: %s with param: %s", query, productID)

	rows, err := postgresDb.Query(query, productID)
	if err != nil {
		log.Println("📢 Error querying variances by product ID:", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}
	defer rows.Close()

	var variances []Variance

	for rows.Next() {
		var v Variance
		if err := rows.Scan(
			&v.ID, &v.ProductName, &v.ProductID, &v.DisplayTitle,
			&v.VarianceDescription, &v.ImageUrl, &v.VarianceTitle,
			&v.Brand, &v.Supplier, &v.OriginalPrice,
			&v.RetailPrice, &v.WholesalePrice,
			&v.Quantity, &v.UnitMeasure, &v.LeastSubUnitMeasure, &v.Barcode,
			&v.CreatedAt, &v.LastModifiedAt,
		); err != nil {
			log.Println("📢 Error scanning row into Variance model:", err)
			continue
		}
		variances = append(variances, v)
	}

	if err = rows.Err(); err != nil {
		log.Println("📢 Error after iterating variances rows:", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to process data returned from database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variances": variances})
}

//! ============================================================================ //
//? ================= ✈️ SUPPLIER RELATED API HANDLERS ✈️ ===================== //
//! ============================================================================ //

func getSupplierFilters(c *gin.Context) {
	query := `
		SELECT 
			COALESCE(id::text, '') AS id, 
			COALESCE(name, '') AS name,
			COALESCE(description, '') AS description,
			COALESCE(logourl, '') AS logourl,
			COALESCE(website, '') AS website,
			COALESCE(created_at, '2025-01-01 00:00:00'::timestamp) AS created_at,
			COALESCE(coutry_of_origin, '') AS coutry_of_origin,
			COALESCE(social_media_links, '') AS social_media_links,
			COALESCE(contact_email, '') AS contact_email,
			COALESCE(phone_number, '') AS phone_number,
			COALESCE(banner_url, '') AS banner_url,
			COALESCE(city, '') AS city,
			COALESCE(country, '') AS country,
			COALESCE(bank_details, '') AS bank_details,
			COALESCE(status, '') AS status,
			COALESCE(extra_data, '') AS extra_data
		FROM supplier_tb
		ORDER BY name ASC;
	`

	rows, err := postgresDb.Query(query)
	if err != nil {
		log.Println("🔴 Failed to fetch suppliers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query suppliers"})
		return
	}
	defer rows.Close()

	var suppliers []Supplier
	for rows.Next() {
		var s Supplier
		err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.LogoURL, &s.Website,
			&s.CreatedAt, &s.CountryOfOrigin, &s.SocialMediaLinks,
			&s.ContactEmail, &s.PhoneNumber, &s.BannerURL,
			&s.LocatedCity, &s.LocatedCountry,
			&s.BankDetails, &s.Status, &s.ExtraData,
		)
		if err != nil {
			log.Println("🔴 Row scan error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse supplier results"})
			return
		}
		suppliers = append(suppliers, s)
	}

	c.JSON(http.StatusOK, gin.H{"suppliers": suppliers})
}

func insertOrUpdateSupplier(c *gin.Context) {
	var supplier Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		log.Println("📢 upserting variances to json parsing got error", err)

		c.JSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_JSON",
					"message": "Invalid JSON input",
					"details": err.Error(),
				},
			})

		return
	}

	now := time.Now()
	supplier.CreatedAt = &now

	query := `
		INSERT INTO supplier_tb (
			name, description, logourl,  coutry_of_origin, 
			social_media_links, contact_email, phone_number, banner_url, 
			website, city, country, bank_details,
			status, extra_data, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (name)
		DO UPDATE SET 
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			logourl = EXCLUDED.logourl,
			coutry_of_origin = EXCLUDED.coutry_of_origin,
			social_media_links = EXCLUDED.social_media_links,
			contact_email = EXCLUDED.contact_email,
			phone_number = EXCLUDED.phone_number,
			banner_url = EXCLUDED.banner_url,
			website = EXCLUDED.website,
			city = EXCLUDED.city,
			country = EXCLUDED.country,
			bank_details = EXCLUDED.bank_details,
			status = EXCLUDED.status,
			extra_data = EXCLUDED.extra_data
		RETURNING id, name, description, logourl, 
			      coutry_of_origin, social_media_links, contact_email, phone_number, 
				  banner_url, website, city, country, 
				  bank_details, status, extra_data, created_at
	`

	// JSON encode the image URL

	var result Supplier

	err := postgresDb.QueryRow(
		query,
		supplier.Name, supplier.Description, supplier.LogoURL, supplier.CountryOfOrigin,
		supplier.SocialMediaLinks, supplier.ContactEmail, supplier.PhoneNumber, supplier.BannerURL,
		supplier.Website, supplier.LocatedCity, supplier.LocatedCountry, supplier.BankDetails,
		supplier.Status, supplier.ExtraData, supplier.CreatedAt,
	).Scan(
		&result.ID, &result.Name, &result.Description, &result.LogoURL, &result.CountryOfOrigin,
		&result.SocialMediaLinks, &result.ContactEmail, &result.PhoneNumber, &result.BannerURL,
		&result.Website, &result.LocatedCity, &result.LocatedCountry, &result.BankDetails,
		&result.Status, &result.ExtraData, &result.CreatedAt,
	)

	if err != nil {
		log.Println("📢 upserting variances to db got error", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "brand upserted",
		"supplier": result,
	})
}

//! ============================================================================ //
//? ================= 🌿 BRAND RELATED API HANDLERS 🌿 ======================== //
//! ============================================================================ //

func getBrandFilters(c *gin.Context) {
	query := `
		SELECT 
			COALESCE(id::text, '') AS id, 
			COALESCE(name, '') AS name,
			COALESCE(description, '') AS description,
			COALESCE(logourl, '') AS logourl,
			COALESCE(created_at, '2025-01-01 00:00:00'::timestamp) AS created_at,
			COALESCE(coutry_of_origin, '') AS coutry_of_origin,
			COALESCE(social_media_links, '') AS social_media_links,
			COALESCE(contact_email, '') AS contact_email,
			COALESCE(phone_number, '') AS phone_number,
			COALESCE(banner_url, '') AS banner_url,
			COALESCE(website, '') AS website
		FROM brand
		ORDER BY name ASC;
	`

	rows, err := postgresDb.Query(query)
	if err != nil {
		log.Println("🔴 Failed to fetch brands:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query brands"})
		return
	}
	defer rows.Close()

	type Brand struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		Description      string `json:"description"`
		LogoURL          string `json:"logourl"`
		CreatedAt        string `json:"created_at"`
		CountryOfOrigin  string `json:"country_of_origin"`
		SocialMediaLinks string `json:"social_media_links"`
		ContactEmail     string `json:"contact_email"`
		PhoneNumber      string `json:"phone_number"`
		BannerURL        string `json:"banner_url"`
		Website          string `json:"website"`
	}

	var brands []Brand
	for rows.Next() {
		var b Brand
		err := rows.Scan(
			&b.ID, &b.Name, &b.Description, &b.LogoURL, &b.CreatedAt,
			&b.CountryOfOrigin, &b.SocialMediaLinks, &b.ContactEmail,
			&b.PhoneNumber, &b.BannerURL, &b.Website,
		)
		if err != nil {
			log.Println("🔴 Row scan error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse brand results"})
			return
		}
		brands = append(brands, b)
	}

	c.JSON(http.StatusOK, gin.H{"brands": brands})
}

func insertOrUpdateBrand(c *gin.Context) {
	var brand Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		log.Println("📢 upserting variances to json parsing got error", err)

		c.JSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_JSON",
					"message": "Invalid JSON input",
					"details": err.Error(),
				},
			})

		return
	}

	now := time.Now()
	brand.CreatedAt = &now

	query := `
		INSERT INTO brand (
			name, description, logourl, coutry_of_origin,
			social_media_links, contact_email, phone_number, 
			banner_url, website, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10
		)
		ON CONFLICT (name)
		DO UPDATE SET 
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			logourl = EXCLUDED.logourl,
			coutry_of_origin = EXCLUDED.coutry_of_origin,
			social_media_links = EXCLUDED.social_media_links,
			contact_email = EXCLUDED.contact_email,
			phone_number = EXCLUDED.phone_number,
			banner_url = EXCLUDED.banner_url,
			website = EXCLUDED.website
		RETURNING id, name, description, logourl, coutry_of_origin,
		          social_media_links, contact_email, phone_number, banner_url, website,
		          created_at
	`

	// JSON encode the image URL

	var result Brand

	err := postgresDb.QueryRow(
		query,
		brand.Name, brand.Description,
		brand.Logourl, brand.CountryOfOrigin, brand.SocialMediaLinks, brand.ContactEmail, brand.PhoneNumber,
		brand.BannerUrl, brand.Website, brand.CreatedAt,
	).Scan(
		&result.ID, &result.Name, &result.Description, &result.Logourl,
		&result.CountryOfOrigin, &result.SocialMediaLinks, &result.ContactEmail, &result.PhoneNumber,
		&result.BannerUrl, &result.Website, &result.CreatedAt,
	)

	if err != nil {
		log.Println("📢 upserting variances to db got error", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to insert product into database",
					"details": err.Error(),
				},
			})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "brand upserted",
		"brand":  result,
	})
}

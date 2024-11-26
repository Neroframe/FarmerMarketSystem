// models/product.go
package models

import (
	"database/sql"
	"strings"
	"time"
)

type Product struct {
	ID          int       `json:"id"`
	FarmerID    int       `json:"farmer_id"`
	Name        string    `json:"name"`
	CategoryID  int       `json:"category_id"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Images      []string  `json:"images"`
}

func CreateProduct(db *sql.DB, product *Product) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (farmer_id, name, category_id, price, quantity, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = tx.QueryRow(query,
		product.FarmerID,
		product.Name,
		product.CategoryID,
		product.Price,
		product.Quantity,
		product.Description,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID)
	if err != nil {
		return err
	}

	// Insert images into product_images table
	for i, img := range product.Images {
		imgQuery := `
			INSERT INTO product_images (product_id, image_url, image_order)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(imgQuery, product.ID, img, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetProductByID(db *sql.DB, id int) (*Product, error) {
	var product Product

	err := db.QueryRow(`
		SELECT 
			id, 
			farmer_id, 
			name, 
			category_id, 
			price, 
			quantity, 
			description, 
			is_active, 
			created_at, 
			updated_at
		FROM products
		WHERE id = $1 AND is_active = TRUE
	`, id).Scan(
		&product.ID,
		&product.FarmerID,
		&product.Name,
		&product.CategoryID,
		&product.Price,
		&product.Quantity,
		&product.Description,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Retrieve images
	rows, err := db.Query(`
		SELECT image_url
		FROM product_images
		WHERE product_id = $1
		ORDER BY image_order ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var img string
		if err := rows.Scan(&img); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	product.Images = images

	return &product, nil
}

func GetActiveProducts(db *sql.DB, farmerID int) ([]Product, error) {
	rows, err := db.Query(`
		SELECT id, farmer_id, name, category_id, price, quantity, description, is_active, created_at, updated_at
		FROM products
		WHERE farmer_id = $1 AND is_active = TRUE
	`, farmerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.ID,
			&product.FarmerID,
			&product.Name,
			&product.CategoryID,
			&product.Price,
			&product.Quantity,
			&product.Description,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Retrieve images for the product
		imgRows, err := db.Query(`
			SELECT image_url
			FROM product_images
			WHERE product_id = $1
			ORDER BY image_order ASC
		`, product.ID)
		if err != nil {
			return nil, err
		}

		var images []string
		for imgRows.Next() {
			var img string
			if err := imgRows.Scan(&img); err != nil {
				imgRows.Close()
				return nil, err
			}
			images = append(images, img)
		}
		imgRows.Close()
		product.Images = images

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func UpdateProduct(db *sql.DB, product *Product) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE products
		SET name = $1, category_id = $2, price = $3, quantity = $4, description = $5, is_active = $6, updated_at = $7
		WHERE id = $8 AND farmer_id = $9
	`
	_, err = tx.Exec(query,
		product.Name,
		product.CategoryID,
		product.Price,
		product.Quantity,
		product.Description,
		product.IsActive,
		product.UpdatedAt,
		product.ID,
		product.FarmerID,
	)
	if err != nil {
		return err
	}

	// Delete existing images
	_, err = tx.Exec(`
		DELETE FROM product_images
		WHERE product_id = $1
	`, product.ID)
	if err != nil {
		return err
	}

	// Insert new images
	for i, img := range product.Images {
		_, err = tx.Exec(`
			INSERT INTO product_images (product_id, image_url, image_order)
			VALUES ($1, $2, $3)
		`, product.ID, img, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func DeleteProduct(db *sql.DB, id int, farmerID int) error {
	_, err := db.Exec(`
		DELETE FROM products
		WHERE id = $1 AND farmer_id = $2
	`, id, farmerID)
	return err
}

func GetAllActiveProducts(db *sql.DB) ([]Product, error) {
	rows, err := db.Query(`
		SELECT id, farmer_id, name, category_id, price, quantity, description, is_active, created_at, updated_at
		FROM products
		WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.ID,
			&product.FarmerID,
			&product.Name,
			&product.CategoryID,
			&product.Price,
			&product.Quantity,
			&product.Description,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Retrieve images for the product
		imgRows, err := db.Query(`
			SELECT image_url
			FROM product_images
			WHERE product_id = $1
			ORDER BY image_order ASC
		`, product.ID)
		if err != nil {
			return nil, err
		}

		var images []string
		for imgRows.Next() {
			var img string
			if err := imgRows.Scan(&img); err != nil {
				imgRows.Close()
				return nil, err
			}
			images = append(images, img)
		}
		imgRows.Close()
		product.Images = images

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func GetProductsWithFilters(db *sql.DB, filters map[string]string, limit, offset int) ([]Product, error) {
	// Initialize query parts
	query := `
		SELECT 
			p.id, 
			p.farmer_id, 
			p.name, 
			p.category_id, 
			p.price, 
			p.quantity, 
			p.description, 
			p.is_active, 
			p.created_at, 
			p.updated_at,
			c.name as category_name,
			f.location as farm_location
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		INNER JOIN farmers f ON p.farmer_id = f.id
		WHERE p.is_active = TRUE
	`
	var params []interface{}
	var conditions []string

	// Filter by category
	if category, ok := filters["category"]; ok && category != "" && strings.ToLower(category) != "all" {
		conditions = append(conditions, "LOWER(c.name) = LOWER(?)")
		params = append(params, category)
	}

	// Search by name, category, or farm location
	if search, ok := filters["search"]; ok && search != "" {
		conditions = append(conditions, `(LOWER(p.name) LIKE LOWER(?) OR LOWER(c.name) LIKE LOWER(?) OR LOWER(f.location) LIKE LOWER(?))`)
		searchTerm := "%" + search + "%"
		params = append(params, searchTerm, searchTerm, searchTerm)
	}

	// Append conditions to the query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Sorting
	if sort, ok := filters["sort"]; ok && sort != "" {
		switch sort {
		case "price_asc":
			query += " ORDER BY p.price ASC"
		case "price_desc":
			query += " ORDER BY p.price DESC"
		case "date_asc":
			query += " ORDER BY p.created_at ASC"
		case "date_desc":
			query += " ORDER BY p.created_at DESC"
		default:
			query += " ORDER BY p.created_at DESC"
		}
	} else {
		query += " ORDER BY p.created_at DESC"
	}

	// Pagination
	query += " LIMIT ? OFFSET ?"
	params = append(params, limit, offset)

	// Prepare the query
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute the query
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		var categoryName string
		var farmLocation string

		err := rows.Scan(
			&product.ID,
			&product.FarmerID,
			&product.Name,
			&product.CategoryID,
			&product.Price,
			&product.Quantity,
			&product.Description,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&categoryName,
			&farmLocation,
		)
		if err != nil {
			return nil, err
		}

		// Retrieve images for the product
		images, err := GetProductImages(db, product.ID)
		if err != nil {
			return nil, err
		}
		product.Images = images

		// Add category name and farm location to the product (if needed)
		// You can extend the Product struct to include these fields or create a new struct

		products = append(products, product)
	}

	return products, nil
}

// GetProductImages retrieves images for a given product ID
func GetProductImages(db *sql.DB, productID int) ([]string, error) {
	rows, err := db.Query(`
		SELECT image_url
		FROM product_images
		WHERE product_id = ?
		ORDER BY image_order ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var img string
		if err := rows.Scan(&img); err != nil {
			return nil, err
		}
		images = append(images, strings.TrimSpace(img))
	}

	return images, nil
}

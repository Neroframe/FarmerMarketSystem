// models/product.go
package models

import (
	"database/sql"
	"fmt"
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
	result, err := db.Exec(`
        DELETE FROM products
        WHERE id = $1 AND farmer_id = $2
    `, id, farmerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Delete related images
	_, err = db.Exec(`
        DELETE FROM product_images
        WHERE product_id = $1
    `, id)
	if err != nil {
		return err
	}

	return nil
}

func GetProductsWithFilters(db *sql.DB, filters map[string]string, limit, offset int) ([]Product, error) {
	query := `
		SELECT id, farmer_id, name, category_id, price, quantity, description, is_active, created_at, updated_at
		FROM products
		WHERE is_active = TRUE
	`

	conditions := []string{}
	params := []interface{}{}
	paramCounter := 1

	// Filter by category
	if category, ok := filters["category"]; ok && category != "" && strings.ToLower(category) != "all" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", paramCounter))
		categoryID := getCategoryIDByName(category) // You'll need to implement this function
		params = append(params, categoryID)
		paramCounter++
	}

	// Search by name
	if search, ok := filters["search"]; ok && search != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(name) LIKE LOWER($%d)", paramCounter))
		searchTerm := "%" + search + "%"
		params = append(params, searchTerm)
		paramCounter++
	}

	// Append conditions to the query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Sorting
	if sort, ok := filters["sort"]; ok && sort != "" {
		switch sort {
		case "price_asc":
			query += " ORDER BY price ASC"
		case "price_desc":
			query += " ORDER BY price DESC"
		case "date_asc":
			query += " ORDER BY created_at ASC"
		case "date_desc":
			query += " ORDER BY created_at DESC"
		default:
			query += " ORDER BY created_at DESC"
		}
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCounter, paramCounter+1)
	params = append(params, limit, offset)
	paramCounter += 2

	// Execute the query
	rows, err := db.Query(query, params...)
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

func getCategoryIDByName(categoryName string) int {
	categoryName = strings.ToLower(categoryName)
	switch categoryName {
	case "vegetables":
		return 1
	case "fruits":
		return 2
	case "seeds":
		return 3
	default:
		return 0 // Or handle unknown categories appropriately
	}
}

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

func GetFarmerLowStockProducts(db *sql.DB, farmerID int, threshold int) ([]Product, error) {
    rows, err := db.Query(`
        SELECT id, farmer_id, name, category_id, price, quantity, description, is_active, created_at, updated_at
        FROM products
        WHERE farmer_id = $1 AND quantity <= $2 AND is_active = TRUE
    `, farmerID, threshold)
    if err != nil {
        return nil, fmt.Errorf("GetFarmerLowStockProducts: error executing query: %w", err)
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
            return nil, fmt.Errorf("GetFarmerLowStockProducts: error scanning row: %w", err)
        }

        images, err := GetProductImages(db, product.ID)
        if err != nil {
            return nil, fmt.Errorf("GetFarmerLowStockProducts: error getting images: %w", err)
        }
        product.Images = images

        products = append(products, product)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("GetFarmerLowStockProducts: rows error: %w", err)
    }

    return products, nil
}

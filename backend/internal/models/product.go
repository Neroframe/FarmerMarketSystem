// models/product.go
package models

import (
	"database/sql"
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

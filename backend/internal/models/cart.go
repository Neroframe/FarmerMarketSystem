// models/cart.go
package models

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// CartItem represents an individual item in the cart
type CartItem struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

// GetCartByBuyerID retrieves all cart items for a specific buyer
func GetCartByBuyerID(db *sql.DB, buyerID int) ([]CartItem, error) {
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
			COALESCE(array_agg(pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL), ARRAY[]::VARCHAR[]) AS images,
			ci.quantity
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE ci.buyer_id = $1
		GROUP BY p.id, p.farmer_id, p.name, p.category_id, p.price, p.quantity, p.description, p.is_active, p.created_at, p.updated_at, ci.quantity
	`

	rows, err := db.Query(query, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []CartItem
	for rows.Next() {
		var product Product
		var images pq.StringArray
		var quantity int

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
			&images,
			&quantity,
		)
		if err != nil {
			return nil, err
		}

		product.Images = images
		cartItem := CartItem{
			Product:  product,
			Quantity: quantity,
		}
		cartItems = append(cartItems, cartItem)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cartItems, nil
}

// AddProductToCart adds a product to the buyer's cart or updates the quantity if it already exists
func AddProductToCart(db *sql.DB, buyerID, productID, quantity int) error {
	if quantity < 1 {
		return errors.New("quantity must be at least 1")
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if the product is already in the cart
	var existingQuantity int
	checkQuery := `SELECT quantity FROM cart_items WHERE buyer_id = $1 AND product_id = $2`
	err = tx.QueryRow(checkQuery, buyerID, productID).Scan(&existingQuantity)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// Insert new cart item
		insertQuery := `INSERT INTO cart_items (buyer_id, product_id, quantity) VALUES ($1, $2, $3)`
		_, err := tx.Exec(insertQuery, buyerID, productID, quantity)
		if err != nil {
			return err
		}
	} else {
		// Update existing cart item
		updateQuery := `UPDATE cart_items SET quantity = quantity + $1 WHERE buyer_id = $2 AND product_id = $3`
		_, err := tx.Exec(updateQuery, quantity, buyerID, productID)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// RemoveProductFromCart removes a product from the buyer's cart
func RemoveProductFromCart(db *sql.DB, buyerID, productID int) error {
	query := `DELETE FROM cart_items WHERE buyer_id = $1 AND product_id = $2`
	res, err := db.Exec(query, buyerID, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("product not found in cart")
	}

	return nil
}

// UpdateCartItem updates the quantity of a product in the buyer's cart
func UpdateCartItem(db *sql.DB, buyerID, productID, quantity int) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	if quantity == 0 {
		// Remove the item from the cart
		return RemoveProductFromCart(db, buyerID, productID)
	}

	// Update the quantity
	query := `UPDATE cart_items SET quantity = $1 WHERE buyer_id = $2 AND product_id = $3`
	res, err := db.Exec(query, quantity, buyerID, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("product not found in cart")
	}

	return nil
}

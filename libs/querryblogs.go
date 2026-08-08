package libs

import "context"

// Deprecated: use QueryBlogs(ctx). Kept for backwards-compat with the old typo name.
func QueryBlogsNoCtx() ([]BlogArticle, error) {
	return QueryBlogs(context.Background())
}

// article is kept for compat with any external JSON decode expectations.
// Use BlogArticle instead.
type article = BlogArticle

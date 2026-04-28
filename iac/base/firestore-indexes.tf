resource "google_firestore_index" "recipe_revisions_by_recipe_created_date" {
  project    = local.project
  database   = "(default)"
  collection = "recipe-revisions"

  fields {
    field_path = "recipeId"
    order      = "ASCENDING"
  }

  fields {
    field_path = "createdDate"
    order      = "DESCENDING"
  }

  query_scope = "COLLECTION"
}

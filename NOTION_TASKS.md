# Agrifolio — Sequential Notion Task List (Site Builder)

1. [ ] Create a new project folder and copy this repo as the starting point.
2. [ ] Rename the module in `go.mod` to your new project name.
3. [ ] Update `cmd/apps/main.go` package paths if the module name changed.
4. [ ] Duplicate `.env.example` to `.env` and fill in local values.
5. [ ] Verify prerequisites: Go, Postgres, Redis, Task are installed.
6. [ ] Run `go mod download` and `task run` to confirm a clean boot.

7. [ ] Add DB migrations: sites, pages, sections (JSONB content), products, assets, contact_profiles.
8. [ ] Add domain models: site/page/section/product/asset/contact_profile in `apps/agrifolio/api/model/`.
9. [ ] Add DTOs for request/response payloads in `apps/agrifolio/api/model/*/dto.go`.
10. [ ] Add repository interfaces + queries for each domain in `apps/agrifolio/api/repository/`.
11. [ ] Add service layer orchestration and validation in `apps/agrifolio/api/service/`.

12. [ ] Implement handlers + routes for Site Setup.
13. [ ] POST `/api/site` creates: site + pages + default sections + empty contact_profile.
14. [ ] GET `/api/site` returns site + related data for the logged-in farmer dashboard.
15. [ ] PATCH `/api/site` updates business_name/tagline/location/theme (regen business_key if renaming).

16. [ ] Implement handlers + routes for Sections (generic resource).
17. [ ] GET `/api/sections` returns all sections for the current site.
18. [ ] PATCH `/api/sections/:id` updates title/subtitle/content/is_enabled.
19. [ ] POST `/api/sections/reorder` updates sort_order from array of {id, sort_order}.
20. [ ] Ensure flexible section data lives in sections.content JSONB.

21. [ ] Implement handlers + routes for Products.
22. [ ] POST `/api/products` creates a product.
23. [ ] GET `/api/products` lists products for the site.
24. [ ] PATCH `/api/products/:id` updates product fields.
25. [ ] DELETE `/api/products/:id` removes a product.

26. [ ] Implement handlers + routes for Assets (uploads).
27. [ ] POST `/api/assets` registers an uploaded file URL + metadata.
28. [ ] DELETE `/api/assets/:id` deletes an asset.
29. [ ] POST `/api/sections/:id/assets` attaches an asset with role/sort_order.
30. [ ] DELETE `/api/sections/:id/assets/:assetId?role=background` detaches section asset.
31. [ ] POST `/api/products/:id/assets` attaches an asset to a product.
32. [ ] DELETE `/api/products/:id/assets/:assetId` detaches product asset.

33. [ ] Implement handlers + routes for Contact Profile (single row per site).
34. [ ] GET `/api/contact` returns contact_profile for the site.
35. [ ] PUT `/api/contact` upserts contact_profile.

36. [ ] Register routes in `apps/agrifolio/api/router/` and wire dependencies in `apps/agrifolio/api/router/router.go`.
37. [ ] Add request validation helpers in `internal/validation/` where needed.
38. [ ] Update OpenAPI docs in `static/openapi.json` for new endpoints.

39. [ ] Add unit tests for repositories/services under `internal/testing/`.
40. [ ] Add handler tests using the testing server helpers.
41. [ ] Run `task test` and fix failing cases.

42. [ ] Run `task migrations:up` and verify database state.
43. [ ] Smoke test endpoints with curl or Postman.
44. [ ] Update `README.md` with new domain overview and setup steps.

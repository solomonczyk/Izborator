import { expect, test } from "@playwright/test"
import http from "node:http"
import { URL } from "node:url"

const mockPort = 3999
let server: http.Server

const goodsFacets = {
  domain: "goods",
  facets: [
    { semantic_type: "price", facet_type: "range" },
    { semantic_type: "category", facet_type: "enum" },
    { semantic_type: "location", facet_type: "enum" },
    { semantic_type: "brand", facet_type: "enum", values: ["Acme", "Globex"] },
  ],
}

const servicesFacets = {
  domain: "services",
  facets: [
    { semantic_type: "category", facet_type: "enum" },
    { semantic_type: "location", facet_type: "enum" },
    { semantic_type: "duration", facet_type: "range" },
    { semantic_type: "price", facet_type: "range" },
  ],
}

function sendJson(res: http.ServerResponse, status: number, payload: unknown) {
  res.statusCode = status
  res.setHeader("content-type", "application/json")
  res.end(JSON.stringify(payload))
}

test.beforeAll(async () => {
  server = http.createServer((req, res) => {
    if (!req.url) {
      sendJson(res, 400, { error: "missing url" })
      return
    }

    const url = new URL(req.url, `http://127.0.0.1:${mockPort}`)

    if (url.pathname === "/api/v1/products/facets") {
      const type = url.searchParams.get("type")
      if (type === "goods") {
        sendJson(res, 200, goodsFacets)
        return
      }
      if (type === "services") {
        sendJson(res, 200, servicesFacets)
        return
      }
      sendJson(res, 400, { error: "invalid type" })
      return
    }

    if (url.pathname === "/api/v1/categories/tree") {
      sendJson(res, 200, [])
      return
    }

    if (url.pathname === "/api/v1/cities") {
      sendJson(res, 200, [])
      return
    }

    if (url.pathname === "/api/v1/products/browse") {
      const page = Number.parseInt(url.searchParams.get("page") || "1", 10)
      const perPage = Number.parseInt(url.searchParams.get("per_page") || "20", 10)
      sendJson(res, 200, {
        items: [],
        page,
        per_page: perPage,
        total: 0,
        total_pages: 0,
      })
      return
    }

    sendJson(res, 404, { error: "not found" })
  })

  await new Promise<void>((resolve) => {
    server.listen(mockPort, "127.0.0.1", resolve)
  })
})

test.afterAll(async () => {
  await new Promise<void>((resolve, reject) => {
    server.close((err) => {
      if (err) {
        reject(err)
        return
      }
      resolve()
    })
  })
})

test("catalog facets gate fields by schema", async ({ page }) => {
  await page.goto("/sr/catalog?type=good")
  await page.waitForLoadState("networkidle")
  await expect(page.locator("select#brand")).toBeVisible()
  await expect(page.locator("input#min_duration")).toHaveCount(0)
  await expect(page.locator("input#max_duration")).toHaveCount(0)

  await page.goto("/sr/catalog?type=service")
  await expect(page.locator("select#brand")).toHaveCount(0)
  await expect(page.locator("input#min_duration")).toBeVisible()
  await expect(page.locator("input#max_duration")).toBeVisible()

  await page.goto("/sr/catalog?type=good&brand=Acme")
  await expect(page.locator("select#brand")).toHaveValue("Acme")
})

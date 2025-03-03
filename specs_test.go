package openapigen

import "strings"

var testBuilderParameterExpectedSpecs = strings.ReplaceAll(`
openapi: 3.0.0
info:
  title: ""
  version: ""
security: null
tags: null
paths:
  /items:
    get:
      parameters:
        - $ref: '#/components/parameters/orderByQueryParam'
        - in: query
          name: order_2
          schema:
            items:
              $ref: '#/components/schemas/OrderBy'
            type: array
      responses:
        default:
          description: ""
components:
  parameters:
    orderByQueryParam:
      in: query
      name: order
      schema:
        items:
          $ref: '#/components/schemas/OrderBy'
        type: array
  schemas:
    OrderBy:
      properties:
        field:
          type: string
        order:
          type: string
      required:
        - field
        - order
      type: object
`, "\n", "")

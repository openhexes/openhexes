import React from "react"
import { createRoot } from "react-dom/client"
import { Route } from "react-router"

import "./index.css"

const Layout = React.lazy(() => import("@/components/utils/layout"))
const NotFound = React.lazy(() => import("@/components/utils/not-found"))

createRoot(document.getElementById("root")!).render(
    <Layout>
        <Route path="*" element={<NotFound />} />
    </Layout>,
)

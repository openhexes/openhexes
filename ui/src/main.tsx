import React from "react"
import { createRoot } from "react-dom/client"
import { Route } from "react-router"

import "./index.css"
import "./vars.css"

const Layout = React.lazy(() => import("@/components/utils/layout"))
const NotFound = React.lazy(() => import("@/components/utils/not-found"))
const MapTest = React.lazy(() => import("@/pages/map-test"))

createRoot(document.getElementById("root")!).render(
    <Layout>
        <Route path="/" element={<MapTest />} />
        <Route path="*" element={<NotFound />} />
    </Layout>,
)

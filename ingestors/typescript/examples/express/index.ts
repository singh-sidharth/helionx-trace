import express, { type Request, type Response } from "express";
import {
  createHelionxExpressMiddleware,
  type HelionxRequestContext,
} from "../../dist/index.js";

const app = express();
app.use(express.json());

app.use(
  createHelionxExpressMiddleware({
    endpoint: "http://localhost:8080",
    service: "express-service",
  }),
);

type HelionxExpressRequest = Request & {
  helionx?: HelionxRequestContext;
};

app.post("/test", async (req: HelionxExpressRequest, res: Response) => {
  try {
    await req.helionx?.success("handler.completed", {
      path: "/test",
      method: req.method,
    });

    res.json({ ok: true });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: "failed to ingest" });
  }
});

app.listen(3000, () => {
  console.log("Example app running on http://localhost:3000");
});
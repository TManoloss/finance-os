import axios from "axios";

const apiServer = axios.create({
  baseURL: `${process.env.NEXT_PUBLIC_API_URL}/api/v1`,
});

export { apiServer };
export default apiServer;

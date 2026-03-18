# saldivia/mcp_server.py
"""MCP Server for RAG Saldivia."""
import asyncio
from mcp.server import Server
from mcp.types import Tool, TextContent
from saldivia.collections import CollectionManager
from saldivia.ingestion_queue import IngestionQueue
import httpx

server = Server("rag-saldivia")


@server.list_tools()
async def list_tools():
    return [
        Tool(
            name="search_documents",
            description="Search documents in a RAG collection",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "Search query"},
                    "collection": {"type": "string", "description": "Collection name"},
                    "top_k": {"type": "integer", "description": "Number of results", "default": 10},
                },
                "required": ["query", "collection"],
            },
        ),
        Tool(
            name="ask_question",
            description="Ask a question using RAG with cross-document synthesis",
            inputSchema={
                "type": "object",
                "properties": {
                    "question": {"type": "string", "description": "Question to answer"},
                    "collection": {"type": "string", "description": "Collection name"},
                },
                "required": ["question", "collection"],
            },
        ),
        Tool(
            name="list_collections",
            description="List all document collections",
            inputSchema={"type": "object", "properties": {}},
        ),
        Tool(
            name="collection_stats",
            description="Get statistics for a collection",
            inputSchema={
                "type": "object",
                "properties": {
                    "collection": {"type": "string", "description": "Collection name"},
                },
                "required": ["collection"],
            },
        ),
        Tool(
            name="ingest_document",
            description="Queue a document for ingestion",
            inputSchema={
                "type": "object",
                "properties": {
                    "file_path": {"type": "string", "description": "Path to document"},
                    "collection": {"type": "string", "description": "Target collection"},
                },
                "required": ["file_path", "collection"],
            },
        ),
        Tool(
            name="ingestion_status",
            description="Check ingestion queue status",
            inputSchema={"type": "object", "properties": {}},
        ),
    ]


@server.call_tool()
async def call_tool(name: str, arguments: dict):
    if name == "search_documents":
        return await search_documents(**arguments)
    elif name == "ask_question":
        return await ask_question(**arguments)
    elif name == "list_collections":
        return await list_collections_tool()
    elif name == "collection_stats":
        return await collection_stats_tool(**arguments)
    elif name == "ingest_document":
        return await ingest_document(**arguments)
    elif name == "ingestion_status":
        return await ingestion_status()
    else:
        raise ValueError(f"Unknown tool: {name}")


async def search_documents(query: str, collection: str, top_k: int = 10):
    """Search documents via RAG API."""
    async with httpx.AsyncClient(timeout=60) as client:
        resp = await client.post(
            "http://localhost:8081/v1/search",
            json={
                "query": query,
                "collection_names": [collection],
                "top_k": top_k,
            }
        )
        results = resp.json()
        return [TextContent(type="text", text=str(results))]


async def ask_question(question: str, collection: str):
    """Answer question via RAG API with streaming."""
    async with httpx.AsyncClient(timeout=120) as client:
        resp = await client.post(
            "http://localhost:8081/v1/generate",
            json={
                "messages": [{"role": "user", "content": question}],
                "collection_names": [collection],
                "use_knowledge_base": True,
            }
        )
        return [TextContent(type="text", text=resp.text)]


async def list_collections_tool():
    """List all collections."""
    manager = CollectionManager()
    collections = manager.list()
    result = "\n".join(f"- {c}" for c in collections) if collections else "No collections"
    return [TextContent(type="text", text=result)]


async def collection_stats_tool(collection: str):
    """Get collection stats."""
    manager = CollectionManager()
    stats = manager.stats(collection)
    if not stats:
        return [TextContent(type="text", text=f"Collection '{collection}' not found")]
    result = f"""Collection: {stats.name}
Entities: {stats.entity_count}
Index: {stats.index_type}
Hybrid: {stats.has_sparse}"""
    return [TextContent(type="text", text=result)]


async def ingest_document(file_path: str, collection: str):
    """Queue document for ingestion."""
    queue = IngestionQueue()
    job = queue.enqueue(file_path, collection)
    return [TextContent(type="text", text=f"Queued job {job.id} for {file_path}")]


async def ingestion_status():
    """Get ingestion queue status."""
    queue = IngestionQueue()
    pending = queue.pending_count()
    jobs = queue.list_jobs()[:5]
    lines = [f"Pending: {pending}"]
    for job in jobs:
        lines.append(f"- {job.id}: {job.status} - {job.file_path}")
    return [TextContent(type="text", text="\n".join(lines))]


def main():
    """Run MCP server."""
    import sys
    from mcp.server.stdio import stdio_server

    async def run():
        async with stdio_server() as (read, write):
            await server.run(read, write, server.create_initialization_options())

    asyncio.run(run())


if __name__ == "__main__":
    main()

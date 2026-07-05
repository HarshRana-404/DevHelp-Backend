package curlconverter

import (
	"fmt"
	"strings"

	"devhelp/internal/services/dto"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ── Go ──────────────────────────────────────────────────────────────────────

type goConverter struct{}

func (g *goConverter) Name() string { return "go" }
func (g *goConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n\t\"fmt\"\n\t\"net/http\"\n\t\"strings\"\n)\n\n")
	b.WriteString("func main() {\n")

	if p.Body != "" {
		body := strings.ReplaceAll(p.Body, "`", "`+\"`\"+`")
		b.WriteString(fmt.Sprintf("\tbody := strings.NewReader(`%s`)\n", body))
		b.WriteString(fmt.Sprintf("\treq, _ := http.NewRequest(%q, %q, body)\n", p.Method, p.URL))
	} else {
		b.WriteString(fmt.Sprintf("\treq, _ := http.NewRequest(%q, %q, nil)\n", p.Method, p.URL))
	}

	for k, v := range p.Headers {
		b.WriteString(fmt.Sprintf("\treq.Header.Set(%q, %q)\n", k, v))
	}

	b.WriteString("\n\tclient := &http.Client{}\n")
	b.WriteString("\tresp, err := client.Do(req)\n")
	b.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n")
	b.WriteString("\tdefer resp.Body.Close()\n")
	b.WriteString("\tfmt.Println(resp.Status)\n")
	b.WriteString("}\n")

	return b.String()
}

// ── Python ──────────────────────────────────────────────────────────────────

type pythonConverter struct{}

func (p *pythonConverter) Name() string { return "python" }
func (p *pythonConverter) Convert(parsed *dto.ParsedCurl) string {
	var b strings.Builder
	caser := cases.Title(language.English)

	b.WriteString("import requests\n\n")

	if len(parsed.Headers) > 0 {
		b.WriteString("headers = {\n")
		for k, v := range parsed.Headers {
			b.WriteString(fmt.Sprintf("    %q: %q,\n", k, v))
		}
		b.WriteString("}\n\n")
	} else {
		b.WriteString("headers = {}\n\n")
	}

	method := caser.String(strings.ToLower(parsed.Method))
	if parsed.Body != "" {
		escaped := strings.ReplaceAll(parsed.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("data = \"%s\"\n\n", escaped))
		b.WriteString(fmt.Sprintf("response = requests.%s(\n    %q,\n    headers=headers,\n    data=data\n)\n", method, parsed.URL))
	} else {
		b.WriteString(fmt.Sprintf("response = requests.%s(\n    %q,\n    headers=headers\n)\n", method, parsed.URL))
	}

	b.WriteString("\nprint(response.status_code)\n")
	b.WriteString("print(response.text)\n")

	return b.String()
}

// ── Java ─────────────────────────────────────────────────────────────────────

type javaConverter struct{}

func (j *javaConverter) Name() string { return "java" }
func (j *javaConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("import java.net.URI;\n")
	b.WriteString("import java.net.http.*;\n")
	b.WriteString("import java.net.http.HttpRequest.BodyPublishers;\n\n")
	b.WriteString("public class Main {\n")
	b.WriteString("    public static void main(String[] args) throws Exception {\n")
	b.WriteString("        HttpClient client = HttpClient.newHttpClient();\n\n")
	b.WriteString(fmt.Sprintf("        HttpRequest.Builder builder = HttpRequest.newBuilder()\n            .uri(URI.create(%q))\n", p.URL))

	for k, v := range p.Headers {
		b.WriteString(fmt.Sprintf("            .header(%q, %q)\n", k, v))
	}

	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("            .method(%q, BodyPublishers.ofString(\"%s\"));\n", p.Method, escaped))
	} else {
		b.WriteString(fmt.Sprintf("            .method(%q, BodyPublishers.noBody());\n", p.Method))
	}

	b.WriteString("\n        HttpResponse<String> response = client.send(\n")
	b.WriteString("            builder.build(),\n")
	b.WriteString("            HttpResponse.BodyHandlers.ofString()\n")
	b.WriteString("        );\n")
	b.WriteString("        System.out.println(response.statusCode());\n")
	b.WriteString("        System.out.println(response.body());\n")
	b.WriteString("    }\n}\n")

	return b.String()
}

// ── C ────────────────────────────────────────────────────────────────────────

type cConverter struct{}

func (c *cConverter) Name() string { return "c" }
func (c *cConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("#include <stdio.h>\n#include <curl/curl.h>\n\n")
	b.WriteString("int main(void) {\n")
	b.WriteString("    CURL *curl = curl_easy_init();\n")
	b.WriteString("    if (!curl) return 1;\n\n")
	b.WriteString(fmt.Sprintf("    curl_easy_setopt(curl, CURLOPT_URL, %q);\n", p.URL))
	b.WriteString(fmt.Sprintf("    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, %q);\n", p.Method))

	if len(p.Headers) > 0 {
		b.WriteString("\n    struct curl_slist *headers = NULL;\n")
		for k, v := range p.Headers {
			b.WriteString(fmt.Sprintf("    headers = curl_slist_append(headers, %q);\n", k+": "+v))
		}
		b.WriteString("    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);\n")
	}

	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("\n    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, \"%s\");\n", escaped))
	}

	b.WriteString("\n    CURLcode res = curl_easy_perform(curl);\n")
	b.WriteString("    curl_easy_cleanup(curl);\n")
	b.WriteString("    return res;\n}\n")

	return b.String()
}

// ── C++ ──────────────────────────────────────────────────────────────────────

type cppConverter struct{}

func (c *cppConverter) Name() string { return "c++" }
func (c *cppConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("#include <iostream>\n#include <curl/curl.h>\n\n")
	b.WriteString("int main() {\n")
	b.WriteString("    CURL* curl = curl_easy_init();\n")
	b.WriteString("    if (!curl) return 1;\n\n")
	b.WriteString(fmt.Sprintf("    curl_easy_setopt(curl, CURLOPT_URL, %q);\n", p.URL))
	b.WriteString(fmt.Sprintf("    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, %q);\n", p.Method))

	if len(p.Headers) > 0 {
		b.WriteString("\n    struct curl_slist* headers = nullptr;\n")
		for k, v := range p.Headers {
			b.WriteString(fmt.Sprintf("    headers = curl_slist_append(headers, %q);\n", k+": "+v))
		}
		b.WriteString("    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);\n")
	}

	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("\n    std::string body = \"%s\";\n", escaped))
		b.WriteString("    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body.c_str());\n")
	}

	b.WriteString("\n    CURLcode res = curl_easy_perform(curl);\n")
	b.WriteString("    if (res != CURLE_OK)\n")
	b.WriteString("        std::cerr << curl_easy_strerror(res) << std::endl;\n")
	b.WriteString("    curl_easy_cleanup(curl);\n")
	b.WriteString("    return 0;\n}\n")

	return b.String()
}

// ── Ruby ─────────────────────────────────────────────────────────────────────

type rubyConverter struct{}

func (r *rubyConverter) Name() string { return "ruby" }
func (r *rubyConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder
	caser := cases.Title(language.English)

	b.WriteString("require 'net/http'\nrequire 'uri'\nrequire 'json'\n\n")
	b.WriteString(fmt.Sprintf("uri = URI.parse(%q)\n", p.URL))
	b.WriteString("http = Net::HTTP.new(uri.host, uri.port)\n")
	b.WriteString("http.use_ssl = uri.scheme == 'https'\n\n")
	b.WriteString(fmt.Sprintf("request = Net::HTTP::%s.new(uri.request_uri)\n", caser.String(strings.ToLower(p.Method))))

	for k, v := range p.Headers {
		b.WriteString(fmt.Sprintf("request[%q] = %q\n", k, v))
	}

	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("\nrequest.body = \"%s\"\n", escaped))
	}

	b.WriteString("\nresponse = http.request(request)\n")
	b.WriteString("puts response.code\nputs response.body\n")

	return b.String()
}

// ── JavaScript (fetch) ───────────────────────────────────────────────────────

type jsConverter struct{}

func (j *jsConverter) Name() string { return "javascript" }
func (j *jsConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("const headers = {\n")
	for k, v := range p.Headers {
		b.WriteString(fmt.Sprintf("  %q: %q,\n", k, v))
	}
	b.WriteString("};\n\n")

	b.WriteString("const options = {\n")
	b.WriteString(fmt.Sprintf("  method: %q,\n", p.Method))
	b.WriteString("  headers,\n")
	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, "`", "\\`")
		b.WriteString(fmt.Sprintf("  body: `%s`,\n", escaped))
	}
	b.WriteString("};\n\n")

	b.WriteString(fmt.Sprintf("fetch(%q, options)\n", p.URL))
	b.WriteString("  .then(res => res.text())\n")
	b.WriteString("  .then(data => console.log(data))\n")
	b.WriteString("  .catch(err => console.error(err));\n")

	return b.String()
}

// ── Kotlin ───────────────────────────────────────────────────────────────────

type kotlinConverter struct{}

func (k *kotlinConverter) Name() string { return "kotlin" }
func (k *kotlinConverter) Convert(p *dto.ParsedCurl) string {
	var b strings.Builder

	b.WriteString("import java.net.URI\nimport java.net.http.*\nimport java.net.http.HttpRequest.BodyPublishers\n\n")
	b.WriteString("fun main() {\n")
	b.WriteString("    val client = HttpClient.newHttpClient()\n\n")
	b.WriteString(fmt.Sprintf("    val request = HttpRequest.newBuilder()\n        .uri(URI.create(%q))\n", p.URL))

	for hk, hv := range p.Headers {
		b.WriteString(fmt.Sprintf("        .header(%q, %q)\n", hk, hv))
	}

	if p.Body != "" {
		escaped := strings.ReplaceAll(p.Body, `"`, `\"`)
		b.WriteString(fmt.Sprintf("        .method(%q, BodyPublishers.ofString(\"%s\"))\n", p.Method, escaped))
	} else {
		b.WriteString(fmt.Sprintf("        .method(%q, BodyPublishers.noBody())\n", p.Method))
	}

	b.WriteString("        .build()\n\n")
	b.WriteString("    val response = client.send(request, HttpResponse.BodyHandlers.ofString())\n")
	b.WriteString("    println(response.statusCode())\n")
	b.WriteString("    println(response.body())\n")
	b.WriteString("}\n")

	return b.String()
}

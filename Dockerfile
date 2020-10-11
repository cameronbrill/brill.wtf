FROM golang:1.14

ENV DB_HOST     = ""
ENV DB_PORT     = "" 
ENV DB_USER     = ""
ENV DB_PASSWORD = ""
ENV DB_NAME     = ""
ENV PORT        = "8080"

EXPOSE 8080

CMD ["go", "run", "main.go"]
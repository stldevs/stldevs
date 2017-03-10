FROM scratch
ADD ./stldevs .
CMD ["./stldevs"]
EXPOSE 8081

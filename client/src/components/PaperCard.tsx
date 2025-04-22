import React from "react";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Paper } from "@/types/paper";

interface PaperCardProps {
  paper: Paper;
}

const PaperCard: React.FC<PaperCardProps> = ({ paper }) => {
  console.log("PaperCard", paper);

  //arxiv.org/abs/physics/0403017v1

  http: return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle className="line-clamp-2 text-xl">{paper.title}</CardTitle>
        <div className="flex justify-normal gap-1 w-full">
          {paper.categories?.map((category, index) => (
            <div
              key={index}
              className="text-sm text-gray-500 bg-gray-100 text-center items-center rounded-full px-4 py-1"
            >
              {category}
            </div>
          ))}
        </div>
      </CardHeader>
      <CardContent className="flex text-lg">
        {new Date(paper.published ?? "").toLocaleDateString(undefined, {
          year: "numeric",
          month: "long",
          day: "numeric",
        })}
      </CardContent>
      <CardFooter>
        <Link href={`/paper/${paper.id.split("/").pop()}`} className="w-full">
          <Button variant="outline" className="w-full">
            View Paper
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
};

export default PaperCard;
